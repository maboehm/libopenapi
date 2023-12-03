// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package v3

import (
	"fmt"
	"github.com/pb33f/libopenapi/datamodel"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
	lowv2 "github.com/pb33f/libopenapi/datamodel/low/v2"
	lowv3 "github.com/pb33f/libopenapi/datamodel/low/v3"
	"github.com/pb33f/libopenapi/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var lowDoc *lowv3.Document

func initTest() {
	data, _ := os.ReadFile("../../../test_specs/burgershop.openapi.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
	})
	if err != nil {
		panic("broken something")
	}
}

func BenchmarkNewDocument(b *testing.B) {
	initTest()
	for i := 0; i < b.N; i++ {
		_ = NewDocument(lowDoc)
	}
}

func TestNewDocument_Extensions(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Equal(t, "darkside", h.Extensions["x-something-something"])
}

func TestNewDocument_ExternalDocs(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Equal(t, "https://pb33f.io", h.ExternalDocs.URL)
}

func TestNewDocument_Security(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Security, 1)
	assert.Len(t, h.Security[0].Requirements, 1)
	assert.Len(t, h.Security[0].Requirements["OAuthScheme"], 2)
}

func TestNewDocument_Info(t *testing.T) {
	initTest()
	highDoc := NewDocument(lowDoc)
	assert.Equal(t, "3.1.0", highDoc.Version)
	assert.Equal(t, "Burger Shop", highDoc.Info.Title)
	assert.Equal(t, "https://pb33f.io", highDoc.Info.TermsOfService)
	assert.Equal(t, "pb33f", highDoc.Info.Contact.Name)
	assert.Equal(t, "buckaroo@pb33f.io", highDoc.Info.Contact.Email)
	assert.Equal(t, "https://pb33f.io", highDoc.Info.Contact.URL)
	assert.Equal(t, "pb33f", highDoc.Info.License.Name)
	assert.Equal(t, "https://pb33f.io/made-up", highDoc.Info.License.URL)
	assert.Equal(t, "1.2", highDoc.Info.Version)
	assert.Equal(t, "https://pb33f.io/schema", highDoc.JsonSchemaDialect)

	wentLow := highDoc.GoLow()
	assert.Equal(t, 1, wentLow.Version.ValueNode.Line)
	assert.Equal(t, 3, wentLow.Info.Value.Title.KeyNode.Line)

	wentLower := highDoc.Info.Contact.GoLow()
	assert.Equal(t, 8, wentLower.Name.ValueNode.Line)
	assert.Equal(t, 11, wentLower.Name.ValueNode.Column)

	wentLowAgain := highDoc.Info.GoLow()
	assert.Equal(t, 3, wentLowAgain.Title.ValueNode.Line)
	assert.Equal(t, 10, wentLowAgain.Title.ValueNode.Column)

	wentOnceMore := highDoc.Info.License.GoLow()
	assert.Equal(t, 12, wentOnceMore.Name.ValueNode.Line)
	assert.Equal(t, 11, wentOnceMore.Name.ValueNode.Column)
}

func TestNewDocument_Servers(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Servers, 2)
	assert.Equal(t, "{scheme}://api.pb33f.io", h.Servers[0].URL)
	assert.Equal(t, "this is our main API server, for all fun API things.", h.Servers[0].Description)
	assert.Len(t, h.Servers[0].Variables, 1)
	assert.Equal(t, "https", h.Servers[0].Variables["scheme"].Default)
	assert.Len(t, h.Servers[0].Variables["scheme"].Enum, 2)

	assert.Equal(t, "https://{domain}.{host}.com", h.Servers[1].URL)
	assert.Equal(t, "this is our second API server, for all fun API things.", h.Servers[1].Description)
	assert.Len(t, h.Servers[1].Variables, 2)
	assert.Equal(t, "api", h.Servers[1].Variables["domain"].Default)
	assert.Equal(t, "pb33f.io", h.Servers[1].Variables["host"].Default)

	wentLow := h.GoLow()
	assert.Equal(t, 45, wentLow.Servers.Value[0].Value.Description.KeyNode.Line)
	assert.Equal(t, 5, wentLow.Servers.Value[0].Value.Description.KeyNode.Column)
	assert.Equal(t, 45, wentLow.Servers.Value[0].Value.Description.ValueNode.Line)
	assert.Equal(t, 18, wentLow.Servers.Value[0].Value.Description.ValueNode.Column)

	wentLower := h.Servers[0].GoLow()
	assert.Equal(t, 45, wentLower.Description.ValueNode.Line)
	assert.Equal(t, 18, wentLower.Description.ValueNode.Column)

	wentLowest := h.Servers[0].Variables["scheme"].GoLow()
	assert.Equal(t, 50, wentLowest.Description.ValueNode.Line)
	assert.Equal(t, 22, wentLowest.Description.ValueNode.Column)
}

func TestNewDocument_Tags(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Tags, 2)
	assert.Equal(t, "Burgers", h.Tags[0].Name)
	assert.Equal(t, "All kinds of yummy burgers.", h.Tags[0].Description)
	assert.Equal(t, "Find out more", h.Tags[0].ExternalDocs.Description)
	assert.Equal(t, "https://pb33f.io", h.Tags[0].ExternalDocs.URL)
	assert.Equal(t, "somethingSpecial", h.Tags[0].Extensions["x-internal-ting"])
	assert.Equal(t, int64(1), h.Tags[0].Extensions["x-internal-tong"])
	assert.Equal(t, 1.2, h.Tags[0].Extensions["x-internal-tang"])
	assert.True(t, h.Tags[0].Extensions["x-internal-tung"].(bool))

	wentLow := h.Tags[1].GoLow()
	assert.Equal(t, 39, wentLow.Description.KeyNode.Line)
	assert.Equal(t, 5, wentLow.Description.KeyNode.Column)

	wentLower := h.Tags[0].ExternalDocs.GoLow()
	assert.Equal(t, 23, wentLower.Description.ValueNode.Line)
	assert.Equal(t, 20, wentLower.Description.ValueNode.Column)
}

func TestNewDocument_Webhooks(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Webhooks, 1)
	assert.Equal(t, "Information about a new burger", h.Webhooks["someHook"].Post.RequestBody.Description)
}

func TestNewDocument_Components_Links(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Links, 2)
	assert.Equal(t, "locateBurger", h.Components.Links["LocateBurger"].OperationId)
	assert.Equal(t, "$response.body#/id", h.Components.Links["LocateBurger"].Parameters["burgerId"])

	wentLow := h.Components.Links["LocateBurger"].GoLow()
	assert.Equal(t, 310, wentLow.OperationId.ValueNode.Line)
	assert.Equal(t, 20, wentLow.OperationId.ValueNode.Column)
}

func TestNewDocument_Components_Callbacks(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Callbacks, 1)
	assert.Equal(
		t,
		"Callback payload",
		h.Components.Callbacks["BurgerCallback"].Expression["{$request.query.queryUrl}"].Post.RequestBody.Description,
	)
	assert.Equal(
		t,
		298,
		h.Components.Callbacks["BurgerCallback"].GoLow().FindExpression("{$request.query.queryUrl}").ValueNode.Line,
	)
	assert.Equal(
		t,
		9,
		h.Components.Callbacks["BurgerCallback"].GoLow().FindExpression("{$request.query.queryUrl}").ValueNode.Column,
	)

	assert.Equal(t, "please", h.Components.Callbacks["BurgerCallback"].Extensions["x-break-everything"])

	for k := range h.Components.GoLow().Callbacks.Value {
		if k.Value == "BurgerCallback" {
			assert.Equal(t, 295, k.KeyNode.Line)
			assert.Equal(t, 5, k.KeyNode.Column)
		}
	}
}

func TestNewDocument_Components_Schemas(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Schemas, 6)

	goLow := h.Components.GoLow()

	a := h.Components.Schemas["Error"]
	abcd := a.Schema().Properties["message"].Schema().Example
	assert.Equal(t, "No such burger as 'Big-Whopper'", abcd)
	assert.Equal(t, 433, goLow.Schemas.KeyNode.Line)
	assert.Equal(t, 3, goLow.Schemas.KeyNode.Column)
	assert.Equal(t, 436, a.Schema().GoLow().Description.KeyNode.Line)

	b := h.Components.Schemas["Burger"]
	assert.Len(t, b.Schema().Required, 2)
	assert.Equal(t, "golden slices of happy fun joy", b.Schema().Properties["fries"].Schema().Description)
	assert.Equal(t, int64(2), b.Schema().Properties["numPatties"].Schema().Example)
	assert.Equal(t, 448, goLow.FindSchema("Burger").Value.Schema().Properties.KeyNode.Line)
	assert.Equal(t, 7, goLow.FindSchema("Burger").Value.Schema().Properties.KeyNode.Column)
	assert.Equal(t, 450, b.Schema().GoLow().FindProperty("name").ValueNode.Line)

	f := h.Components.Schemas["Fries"]
	assert.Equal(t, "salt", f.Schema().Properties["seasoning"].Schema().Items.A.Schema().Example)
	assert.Len(t, f.Schema().Properties["favoriteDrink"].Schema().Properties["drinkType"].Schema().Enum, 1)

	d := h.Components.Schemas["Drink"]
	assert.Len(t, d.Schema().Required, 2)
	assert.True(t, d.Schema().AdditionalProperties.B)
	assert.Equal(t, "drinkType", d.Schema().Discriminator.PropertyName)
	assert.Equal(t, "some value", d.Schema().Discriminator.Mapping["drink"])
	assert.Equal(t, 516, d.Schema().Discriminator.GoLow().PropertyName.ValueNode.Line)
	assert.Equal(t, 23, d.Schema().Discriminator.GoLow().PropertyName.ValueNode.Column)

	pl := h.Components.Schemas["SomePayload"]
	assert.Equal(t, "is html programming? yes.", pl.Schema().XML.Name)
	assert.Equal(t, 523, pl.Schema().XML.GoLow().Name.ValueNode.Line)

	ext := h.Components.Extensions
	assert.Equal(t, "loud", ext["x-screaming-baby"])
}

func TestNewDocument_Components_Headers(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Headers, 1)
	assert.Equal(t, "this is a header example for UseOil", h.Components.Headers["UseOil"].Description)
	assert.Equal(t, 323, h.Components.Headers["UseOil"].GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, h.Components.Headers["UseOil"].GoLow().Description.ValueNode.Column)
}

func TestNewDocument_Components_RequestBodies(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.RequestBodies, 1)
	assert.Equal(t, "Give us the new burger!", h.Components.RequestBodies["BurgerRequest"].Description)
	assert.Equal(t, 328, h.Components.RequestBodies["BurgerRequest"].GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, h.Components.RequestBodies["BurgerRequest"].GoLow().Description.ValueNode.Column)
	assert.Len(t, h.Components.RequestBodies["BurgerRequest"].Content["application/json"].Examples, 2)
}

func TestNewDocument_Components_Examples(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Examples, 1)
	assert.Equal(t, "A juicy two hander sammich", h.Components.Examples["QuarterPounder"].Summary)
	assert.Equal(t, 346, h.Components.Examples["QuarterPounder"].GoLow().Summary.ValueNode.Line)
	assert.Equal(t, 16, h.Components.Examples["QuarterPounder"].GoLow().Summary.ValueNode.Column)
}

func TestNewDocument_Components_Responses(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Responses, 1)
	assert.Equal(t, "all the dressings for a burger.", h.Components.Responses["DressingResponse"].Description)
	assert.Equal(t, "array", h.Components.Responses["DressingResponse"].Content["application/json"].Schema.Schema().Type[0])
	assert.Equal(t, 352, h.Components.Responses["DressingResponse"].GoLow().Description.KeyNode.Line)
	assert.Equal(t, 7, h.Components.Responses["DressingResponse"].GoLow().Description.KeyNode.Column)
}

func TestNewDocument_Components_SecuritySchemes(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.SecuritySchemes, 3)

	api := h.Components.SecuritySchemes["APIKeyScheme"]
	assert.Equal(t, "an apiKey security scheme", api.Description)
	assert.Equal(t, 364, api.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, api.GoLow().Description.ValueNode.Column)

	jwt := h.Components.SecuritySchemes["JWTScheme"]
	assert.Equal(t, "an JWT security scheme", jwt.Description)
	assert.Equal(t, 369, jwt.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, jwt.GoLow().Description.ValueNode.Column)

	oAuth := h.Components.SecuritySchemes["OAuthScheme"]
	assert.Equal(t, "an oAuth security scheme", oAuth.Description)
	assert.Equal(t, 375, oAuth.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, oAuth.GoLow().Description.ValueNode.Column)
	assert.Len(t, oAuth.Flows.Implicit.Scopes, 2)
	assert.Equal(t, "read all burgers", oAuth.Flows.Implicit.Scopes["read:burgers"])
	assert.Equal(t, "https://pb33f.io/oauth", oAuth.Flows.AuthorizationCode.AuthorizationUrl)

	// check the lowness is low.
	assert.Equal(t, 380, oAuth.Flows.GoLow().Implicit.Value.Scopes.KeyNode.Line)
	assert.Equal(t, 11, oAuth.Flows.GoLow().Implicit.Value.Scopes.KeyNode.Column)
	assert.Equal(t, 380, oAuth.Flows.Implicit.GoLow().Scopes.KeyNode.Line)
	assert.Equal(t, 11, oAuth.Flows.Implicit.GoLow().Scopes.KeyNode.Column)
}

func TestNewDocument_Components_Parameters(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Components.Parameters, 2)
	bh := h.Components.Parameters["BurgerHeader"]
	assert.Equal(t, "burgerHeader", bh.Name)
	assert.Equal(t, 392, bh.GoLow().Name.KeyNode.Line)
	assert.Len(t, bh.Schema.Schema().Properties, 2)
	assert.Equal(t, "big-mac", bh.Example)
	assert.True(t, bh.Required)
	assert.Equal(
		t,
		"this is a header",
		bh.Content["application/json"].Encoding["burgerTheme"].Headers["someHeader"].Description,
	)
	assert.Len(t, bh.Content["application/json"].Schema.Schema().Properties, 2)
	assert.Equal(t, 409, bh.Content["application/json"].Encoding["burgerTheme"].GoLow().ContentType.ValueNode.Line)
}

func TestNewDocument_Paths(t *testing.T) {
	initTest()
	h := NewDocument(lowDoc)
	assert.Len(t, h.Paths.PathItems, 5)

	testBurgerShop(t, h, true)
}

func testBurgerShop(t *testing.T, h *Document, checkLines bool) {
	burgersOp := h.Paths.PathItems["/burgers"]

	assert.Len(t, burgersOp.GetOperations(), 1)
	assert.Equal(t, "meaty", burgersOp.Extensions["x-burger-meta"])
	assert.Nil(t, burgersOp.Get)
	assert.Nil(t, burgersOp.Put)
	assert.Nil(t, burgersOp.Patch)
	assert.Nil(t, burgersOp.Head)
	assert.Nil(t, burgersOp.Options)
	assert.Nil(t, burgersOp.Trace)

	assert.Equal(t, "createBurger", burgersOp.Post.OperationId)
	assert.Len(t, burgersOp.Post.Tags, 1)
	assert.Equal(t, "A new burger for our menu, yummy yum yum.", burgersOp.Post.Description)
	assert.Equal(t, "Give us the new burger!", burgersOp.Post.RequestBody.Description)
	assert.Len(t, burgersOp.Post.Responses.Codes, 3)
	if checkLines {
		assert.Equal(t, 64, burgersOp.GoLow().Post.KeyNode.Line)
		assert.Equal(t, 63, h.Paths.GoLow().FindPath("/burgers").ValueNode.Line)
	}

	okResp := burgersOp.Post.Responses.FindResponseByCode(200)
	assert.Len(t, okResp.Headers, 1)
	assert.Equal(t, "A tasty burger for you to eat.", okResp.Description)
	assert.Len(t, okResp.Content["application/json"].Examples, 2)
	assert.Equal(
		t,
		"a cripsy fish sammich filled with ocean goodness.",
		okResp.Content["application/json"].Examples["filetOFish"].Summary,
	)
	assert.Len(t, okResp.Links, 2)
	assert.Equal(t, "locateBurger", okResp.Links["LocateBurger"].OperationId)
	assert.Len(t, burgersOp.Post.Security[0].Requirements, 1)
	assert.Len(t, burgersOp.Post.Security[0].Requirements["OAuthScheme"], 2)
	assert.Equal(t, "read:burgers", burgersOp.Post.Security[0].Requirements["OAuthScheme"][0])
	assert.Len(t, burgersOp.Post.Servers, 1)
	assert.Equal(t, "https://pb33f.io", burgersOp.Post.Servers[0].URL)

	if checkLines {
		assert.Equal(t, 69, burgersOp.Post.GoLow().Description.ValueNode.Line)
		assert.Equal(t, 74, burgersOp.Post.Responses.GoLow().FindResponseByCode("200").ValueNode.Line)
		assert.Equal(t, 80, okResp.Content["application/json"].GoLow().Schema.KeyNode.Line)
		assert.Equal(t, 15, okResp.Content["application/json"].GoLow().Schema.KeyNode.Column)
		assert.Equal(t, 77, okResp.GoLow().Description.KeyNode.Line)
		assert.Equal(t, 310, okResp.Links["LocateBurger"].GoLow().OperationId.ValueNode.Line)
		assert.Equal(t, 118, burgersOp.Post.Security[0].GoLow().Requirements.ValueNode.Line)
	}
}

func TestStripeAsDoc(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/stripe.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	assert.Len(t, utils.UnwrapErrors(err), 1)
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
}

func TestK8sAsDoc(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/k8s.json")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error
	lowSwag, err := lowv2.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	d := v2.NewSwaggerDocument(lowSwag)
	assert.Len(t, utils.UnwrapErrors(err), 0)
	assert.NotNil(t, d)
}

func TestAsanaAsDoc(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/asana.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	if err != nil {
		panic("broken something")
	}
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
	assert.Equal(t, 118, len(d.Paths.PathItems))
}

func TestDigitalOceanAsDocViaCheckout(t *testing.T) {

	// this is a full checkout of the digitalocean API repo.
	tmp, _ := os.MkdirTemp("", "openapi")
	cmd := exec.Command("git", "clone", "https://github.com/digitalocean/openapi", tmp)
	defer os.RemoveAll(filepath.Join(tmp, "openapi"))

	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	spec, _ := filepath.Abs(filepath.Join(tmp, "specification", "DigitalOcean-public.v2.yaml"))
	doLocal, _ := os.ReadFile(spec)

	var rootNode yaml.Node
	_ = yaml.Unmarshal(doLocal, &rootNode)

	basePath := filepath.Join(tmp, "specification")

	data, _ := os.ReadFile("../../../test_specs/digitalocean.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)

	config := datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BasePath:              basePath,
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &config)
	if err != nil {
		er := utils.UnwrapErrors(err)
		for e := range er {
			fmt.Println(er[e])
		}
	}
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
	assert.Equal(t, 183, len(d.Paths.PathItems))
}

func TestDigitalOceanAsDocFromSHA(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/digitalocean.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error

	baseURL, _ := url.Parse("https://raw.githubusercontent.com/digitalocean/openapi/82e1d558e15a59edc1d47d2c5544e7138f5b3cbf/specification")
	config := datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BaseURL:               baseURL,
	}

	if os.Getenv("GH_PAT") != "" {
		client := &http.Client{
			Timeout: time.Second * 60,
		}
		config.RemoteURLHandler = func(url string) (*http.Response, error) {
			request, _ := http.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GH_PAT")))
			return client.Do(request)
		}
	}

	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &config)
	assert.Len(t, utils.UnwrapErrors(err), 3) // there are 3 404's in this release of the API.
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
	assert.Equal(t, 183, len(d.Paths.PathItems))
}

func TestDigitalOceanAsDocFromMain(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/digitalocean.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error

	baseURL, _ := url.Parse("https://raw.githubusercontent.com/digitalocean/openapi/main/specification")
	config := datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BaseURL:               baseURL,
	}

	config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	if os.Getenv("GH_PAT") != "" {
		client := &http.Client{
			Timeout: time.Second * 60,
		}
		config.RemoteURLHandler = func(url string) (*http.Response, error) {
			request, _ := http.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_TOKEN")))
			return client.Do(request)
		}
	}

	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &config)
	if err != nil {
		er := utils.UnwrapErrors(err)
		for e := range er {
			fmt.Printf("Reported Error: %s\n", er[e])
		}
	}
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
	assert.Equal(t, 183, len(d.Paths.PathItems))
}

func TestPetstoreAsDoc(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/petstorev3.json")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	if err != nil {
		panic("broken something")
	}
	d := NewDocument(lowDoc)
	assert.NotNil(t, d)
	assert.Equal(t, 13, len(d.Paths.PathItems))
}

func TestCircularReferencesDoc(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/circular-tests.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)

	lDoc, err := lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	assert.Len(t, utils.UnwrapErrors(err), 3)
	d := NewDocument(lDoc)
	assert.Len(t, d.Components.Schemas, 9)
	assert.Len(t, d.Index.GetCircularReferences(), 3)
}

func TestDocument_MarshalYAML(t *testing.T) {
	// create a new document
	initTest()
	h := NewDocument(lowDoc)

	// render the document to YAML
	r, _ := h.Render()

	info, _ := datamodel.ExtractSpecInfo(r)
	lDoc, e := lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	assert.Nil(t, e)

	highDoc := NewDocument(lDoc)
	testBurgerShop(t, highDoc, false)
}

func TestDocument_MarshalIndention(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/single-definition.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)

	lowDoc, _ = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())

	highDoc := NewDocument(lowDoc)
	rendered := highDoc.RenderWithIndention(2)

	assert.Equal(t, string(data), strings.TrimSpace(string(rendered)))

	rendered = highDoc.RenderWithIndention(4)

	assert.NotEqual(t, string(data), strings.TrimSpace(string(rendered)))
}

func TestDocument_MarshalIndention_Error(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/single-definition.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)

	lowDoc, _ = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())

	highDoc := NewDocument(lowDoc)
	rendered := highDoc.RenderWithIndention(2)

	assert.Equal(t, string(data), strings.TrimSpace(string(rendered)))

	rendered = highDoc.RenderWithIndention(4)

	assert.NotEqual(t, string(data), strings.TrimSpace(string(rendered)))
}

func TestDocument_MarshalJSON(t *testing.T) {
	data, _ := os.ReadFile("../../../test_specs/petstorev3.json")
	info, _ := datamodel.ExtractSpecInfo(data)

	lowDoc, _ = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())

	highDoc := NewDocument(lowDoc)

	rendered := highDoc.RenderJSON("  ")

	// now read back in the JSON
	info, _ = datamodel.ExtractSpecInfo(rendered)
	lowDoc, _ = lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	newDoc := NewDocument(lowDoc)

	assert.Equal(t, len(newDoc.Paths.PathItems), len(highDoc.Paths.PathItems))
	assert.Equal(t, len(newDoc.Components.Schemas), len(highDoc.Components.Schemas))
}

func TestDocument_MarshalYAMLInline(t *testing.T) {
	// create a new document
	initTest()
	h := NewDocument(lowDoc)

	// render the document to YAML inline
	r, _ := h.RenderInline()

	info, _ := datamodel.ExtractSpecInfo(r)
	lDoc, e := lowv3.CreateDocumentFromConfig(info, datamodel.NewDocumentConfiguration())
	assert.Nil(t, e)

	highDoc := NewDocument(lDoc)
	testBurgerShop(t, highDoc, false)
}

func TestDocument_MarshalYAML_TestRefs(t *testing.T) {
	// create a new document
	yml := `openapi: 3.1.0
paths:
    x-milky-milk: milky
    /burgers:
        x-burger-meta: meaty
        post:
            operationId: createBurger
            tags:
                - Burgers
            summary: Create a new burger
            description: A new burger for our menu, yummy yum yum.
            responses:
                "200":
                    headers:
                        UseOil:
                            $ref: '#/components/headers/UseOil'
                    description: A tasty burger for you to eat.
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Burger'
                            examples:
                                quarterPounder:
                                    $ref: '#/components/examples/QuarterPounder'
                                filetOFish:
                                    summary: a cripsy fish sammich filled with ocean goodness.
                                    value:
                                        name: Filet-O-Fish
                                        numPatties: 1
components:
    headers:
        UseOil:
            description: this is a header example for UseOil
            schema:
                type: string
    schemas:
        Burger:
            type: object
            description: The tastiest food on the planet you would love to eat everyday
            required:
                - name
                - numPatties
            properties:
                name:
                    type: string
                    description: The name of your tasty burger - burger names are listed in our menus
                    example: Big Mac
                numPatties:
                    type: integer
                    description: The number of burger patties used
                    example: "2"
    examples:
        QuarterPounder:
            summary: A juicy two hander sammich
            value:
                name: Quarter Pounder with Cheese
                numPatties: 1`

	info, _ := datamodel.ExtractSpecInfo([]byte(yml))
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
	})
	if err != nil {
		panic("broken something")
	}
	h := NewDocument(lowDoc)

	// render the document to YAML, and it should be identical to the original in size, example ordering isn't
	// guaranteed, so we can't compare the strings directly
	r, _ := h.Render()
	assert.Len(t, strings.TrimSpace(string(r)), len(strings.TrimSpace(yml)))
}

func TestDocument_MarshalYAML_TestParamRefs(t *testing.T) {
	// create a new document
	yml := `openapi: 3.1.0
paths:
    /burgers/{burgerId}:
        get:
            operationId: locateBurger
            tags:
                - Burgers
            summary: Search a burger by ID - returns the burger with that identifier
            description: Look up a tasty burger take it and enjoy it
            parameters:
                - $ref: '#/components/parameters/BurgerId'
                - $ref: '#/components/parameters/BurgerHeader'
components:
    parameters:
        BurgerHeader:
            in: header
            name: burgerHeader
            schema:
                properties:
                    burgerTheme:
                        type: string
                        description: something about a theme goes in here?
                    burgerTime:
                        type: number
                        description: number of burgers ordered so far this year.
        BurgerId:
            in: path
            name: burgerId
            schema:
                type: string
            example: big-mac
            description: the name of the burger. use this to order your tasty burger
            required: true`

	info, _ := datamodel.ExtractSpecInfo([]byte(yml))
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
	})
	if err != nil {
		panic("broken something")
	}
	h := NewDocument(lowDoc)

	// render the document to YAML and it should be identical.
	r, _ := h.Render()
	assert.Equal(t, yml, strings.TrimSpace(string(r)))
}

func TestDocument_MarshalYAML_TestModifySchemas(t *testing.T) {
	// create a new document
	yml := `openapi: 3.1.0
components:
  schemas:
    BurgerHeader:
      properties:
        burgerTheme:
          type: string
          description: something about a theme goes in here?
`

	info, _ := datamodel.ExtractSpecInfo([]byte(yml))
	var err error
	lowDoc, err = lowv3.CreateDocumentFromConfig(info, &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
	})
	if err != nil {
		panic("broken something")
	}
	h := NewDocument(lowDoc)

	// mutate the schema
	g := h.Components.Schemas["BurgerHeader"].Schema()
	ds := g.Properties["burgerTheme"].Schema()
	ds.Description = "changed"

	// render the document to YAML and it should be identical.
	r, _ := h.Render()

	desired := `openapi: 3.1.0
components:
    schemas:
        BurgerHeader:
            properties:
                burgerTheme:
                    type: string
                    description: changed`

	assert.Equal(t, desired, strings.TrimSpace(string(r)))
}
