// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package v3

import (
	"github.com/pb33f/libopenapi/datamodel"
	lowv3 "github.com/pb33f/libopenapi/datamodel/low/3.0"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var doc *lowv3.Document

func init() {
	data, _ := ioutil.ReadFile("../../../test_specs/burgershop.openapi.yaml")
	info, _ := datamodel.ExtractSpecInfo(data)
	var err []error
	doc, err = lowv3.CreateDocument(info)
	if err != nil {
		panic("broken something")
	}
}

func BenchmarkNewDocument(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewDocument(doc)
	}
}

func TestNewDocument_Extensions(t *testing.T) {
	h := NewDocument(doc)
	assert.Equal(t, "darkside", h.Extensions["x-something-something"])
}

func TestNewDocument_ExternalDocs(t *testing.T) {
	h := NewDocument(doc)
	assert.Equal(t, "https://pb33f.io", h.ExternalDocs.URL)
}

func TestNewDocument_Info(t *testing.T) {
	highDoc := NewDocument(doc)
	assert.Equal(t, "3.0.1", highDoc.Version)
	assert.Equal(t, "Burger Shop", highDoc.Info.Title)
	assert.Equal(t, "https://pb33f.io", highDoc.Info.TermsOfService)
	assert.Equal(t, "pb33f", highDoc.Info.Contact.Name)
	assert.Equal(t, "buckaroo@pb33f.io", highDoc.Info.Contact.Email)
	assert.Equal(t, "https://pb33f.io", highDoc.Info.Contact.URL)
	assert.Equal(t, "pb33f", highDoc.Info.License.Name)
	assert.Equal(t, "https://pb33f.io/made-up", highDoc.Info.License.URL)
	assert.Equal(t, "1.2", highDoc.Info.Version)

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
	h := NewDocument(doc)
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
	h := NewDocument(doc)
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

func TestNewDocument_Components_Links(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Links, 2)
	assert.Equal(t, "locateBurger", h.Components.Links["LocateBurger"].OperationId)
	assert.Equal(t, "$response.body#/id", h.Components.Links["LocateBurger"].Parameters["burgerId"])

	wentLow := h.Components.Links["LocateBurger"].GoLow()
	assert.Equal(t, 305, wentLow.OperationId.ValueNode.Line)
	assert.Equal(t, 20, wentLow.OperationId.ValueNode.Column)

}

func TestNewDocument_Components_Callbacks(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Callbacks, 1)
	assert.Equal(t, "Callback payload",
		h.Components.Callbacks["BurgerCallback"].Expression["{$request.query.queryUrl}"].Post.RequestBody.Description)

	assert.Equal(t, 293,
		h.Components.Callbacks["BurgerCallback"].GoLow().FindExpression("{$request.query.queryUrl}").ValueNode.Line)
	assert.Equal(t, 9,
		h.Components.Callbacks["BurgerCallback"].GoLow().FindExpression("{$request.query.queryUrl}").ValueNode.Column)

	assert.Equal(t, "please", h.Components.Callbacks["BurgerCallback"].Extensions["x-break-everything"])

	for k := range h.Components.GoLow().Callbacks.Value {
		if k.Value == "BurgerCallback" {
			assert.Equal(t, 290, k.KeyNode.Line)
			assert.Equal(t, 5, k.KeyNode.Column)
		}
	}
}

func TestNewDocument_Components_Schemas(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Schemas, 6)

	goLow := h.Components.GoLow()

	a := h.Components.Schemas["Error"]
	assert.Equal(t, "No such burger as 'Big-Whopper'", a.Properties["message"].Example)
	assert.Equal(t, 428, goLow.Schemas.KeyNode.Line)
	assert.Equal(t, 3, goLow.Schemas.KeyNode.Column)
	assert.Equal(t, 431, a.GoLow().Description.KeyNode.Line)

	b := h.Components.Schemas["Burger"]
	assert.Len(t, b.Required, 2)
	assert.Equal(t, "golden slices of happy fun joy", b.Properties["fries"].Description)
	assert.Equal(t, int64(2), b.Properties["numPatties"].Example)
	assert.Equal(t, 443, goLow.FindSchema("Burger").Value.Properties.KeyNode.Line)
	assert.Equal(t, 7, goLow.FindSchema("Burger").Value.Properties.KeyNode.Column)
	assert.Equal(t, 445, b.GoLow().FindProperty("name").ValueNode.Line)

	f := h.Components.Schemas["Fries"]
	assert.Equal(t, "salt", f.Properties["seasoning"].Items[0].Example)
	assert.Len(t, f.Properties["favoriteDrink"].Properties["drinkType"].Enum, 2)

	d := h.Components.Schemas["Drink"]
	assert.Len(t, d.Required, 2)
	assert.True(t, d.AdditionalProperties.(bool))
	assert.Equal(t, "drinkType", d.Discriminator.PropertyName)
	assert.Equal(t, "some value", d.Discriminator.Mapping["drink"])

	ext := h.Components.Extensions
	assert.Equal(t, "loud", ext["x-screaming-baby"])
}

func TestNewDocument_Components_Headers(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Headers, 1)
	assert.Equal(t, "this is a header", h.Components.Headers["UseOil"].Description)
	assert.Equal(t, 318, h.Components.Headers["UseOil"].GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, h.Components.Headers["UseOil"].GoLow().Description.ValueNode.Column)
}

func TestNewDocument_Components_RequestBodies(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.RequestBodies, 1)
	assert.Equal(t, "Give us the new burger!", h.Components.RequestBodies["BurgerRequest"].Description)
	assert.Equal(t, 323, h.Components.RequestBodies["BurgerRequest"].GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, h.Components.RequestBodies["BurgerRequest"].GoLow().Description.ValueNode.Column)
	assert.Len(t, h.Components.RequestBodies["BurgerRequest"].Content["application/json"].Examples, 2)
}

func TestNewDocument_Components_Examples(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Examples, 1)
	assert.Equal(t, "A juicy two hander sammich", h.Components.Examples["QuarterPounder"].Summary)
	assert.Equal(t, 341, h.Components.Examples["QuarterPounder"].GoLow().Summary.ValueNode.Line)
	assert.Equal(t, 16, h.Components.Examples["QuarterPounder"].GoLow().Summary.ValueNode.Column)

}

func TestNewDocument_Components_Responses(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Responses, 1)
	assert.Equal(t, "all the dressings for a burger.", h.Components.Responses["DressingResponse"].Description)
	assert.Equal(t, "array", h.Components.Responses["DressingResponse"].Content["application/json"].Schema.Type)
	assert.Equal(t, 347, h.Components.Responses["DressingResponse"].GoLow().Description.KeyNode.Line)
	assert.Equal(t, 7, h.Components.Examples["QuarterPounder"].GoLow().Summary.KeyNode.Column)
}

func TestNewDocument_Components_SecuritySchemes(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.SecuritySchemes, 3)

	api := h.Components.SecuritySchemes["APIKeyScheme"]
	assert.Equal(t, "an apiKey security scheme", api.Description)
	assert.Equal(t, 359, api.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, api.GoLow().Description.ValueNode.Column)

	jwt := h.Components.SecuritySchemes["JWTScheme"]
	assert.Equal(t, "an JWT security scheme", jwt.Description)
	assert.Equal(t, 364, jwt.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, jwt.GoLow().Description.ValueNode.Column)

	oAuth := h.Components.SecuritySchemes["OAuthScheme"]
	assert.Equal(t, "an oAuth security scheme", oAuth.Description)
	assert.Equal(t, 370, oAuth.GoLow().Description.ValueNode.Line)
	assert.Equal(t, 20, oAuth.GoLow().Description.ValueNode.Column)
	assert.Len(t, oAuth.Flows.Implicit.Scopes, 2)
	assert.Equal(t, "read all burgers", oAuth.Flows.Implicit.Scopes["read:burgers"])
	assert.Equal(t, "https://pb33f.io/oauth", oAuth.Flows.AuthorizationCode.AuthorizationUrl)

	// check the lowness is low.
	assert.Equal(t, 375, oAuth.Flows.GoLow().Implicit.Value.Scopes.KeyNode.Line)
	assert.Equal(t, 11, oAuth.Flows.GoLow().Implicit.Value.Scopes.KeyNode.Column)
	assert.Equal(t, 375, oAuth.Flows.Implicit.GoLow().Scopes.KeyNode.Line)
	assert.Equal(t, 11, oAuth.Flows.Implicit.GoLow().Scopes.KeyNode.Column)

}

func TestNewDocument_Components_Parameters(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Components.Parameters, 2)
	bh := h.Components.Parameters["BurgerHeader"]
	assert.Equal(t, "burgerHeader", bh.Name)
	assert.Equal(t, 387, bh.GoLow().Name.KeyNode.Line)
	assert.Len(t, bh.Schema.Properties, 2)
	assert.Equal(t, "big-mac", bh.Example)
	assert.True(t, bh.Required)
	assert.Equal(t, "this is a header",
		bh.Content["application/json"].Encoding["burgerTheme"].Headers["someHeader"].Description)
	assert.Len(t, bh.Content["application/json"].Schema.Properties, 2)
}

func TestNewDocument_Paths(t *testing.T) {
	h := NewDocument(doc)
	assert.Len(t, h.Paths.PathItems, 5)

	burgersOp := h.Paths.PathItems["/burgers"]
	assert.Equal(t, "meaty", burgersOp.Extensions["x-burger-meta"])
	assert.Nil(t, burgersOp.Get)
	assert.Nil(t, burgersOp.Put)
	assert.Nil(t, burgersOp.Patch)
	assert.Nil(t, burgersOp.Head)
	assert.Nil(t, burgersOp.Options)
	assert.Nil(t, burgersOp.Trace)
	assert.Equal(t, 64, burgersOp.GoLow().Post.KeyNode.Line)
	assert.Equal(t, "createBurger", burgersOp.Post.OperationId)
	assert.Len(t, burgersOp.Post.Tags, 1)
	assert.Equal(t, "A new burger for our menu, yummy yum yum.", burgersOp.Post.Description)
	assert.Equal(t, "Give us the new burger!", burgersOp.Post.RequestBody.Description)
	assert.Len(t, burgersOp.Post.Responses.Codes, 3)

	okResp := burgersOp.Post.Responses.FindResponseByCode(200)
	assert.Len(t, okResp.Headers, 1)
	assert.Equal(t, "A tasty burger for you to eat.", okResp.Description)
	assert.Equal(t, 69, burgersOp.Post.GoLow().Description.ValueNode.Line)
	assert.Len(t, okResp.Content["application/json"].Examples, 2)
	assert.Equal(t, "a cripsy fish sammich filled with ocean goodness.",
		okResp.Content["application/json"].Examples["filetOFish"].Summary)
	assert.Equal(t, 77, okResp.GoLow().Description.KeyNode.Line)
	assert.Len(t, okResp.Links, 2)
	assert.Equal(t, "locateBurger", okResp.Links["LocateBurger"].OperationId)
	assert.Equal(t, 305, okResp.Links["LocateBurger"].GoLow().OperationId.ValueNode.Line)
	assert.Len(t, burgersOp.Post.Security.ValueRequirements[0], 1)
	assert.Len(t, burgersOp.Post.Security.ValueRequirements[0]["OAuthScheme"], 2)
	assert.Equal(t, "read:burgers", burgersOp.Post.Security.ValueRequirements[0]["OAuthScheme"][0])
	assert.Equal(t, 118, burgersOp.Post.Security.GoLow().ValueRequirements[0].ValueNode.Line)
	assert.Len(t, burgersOp.Post.Servers, 1)
	assert.Equal(t, "https://pb33f.io", burgersOp.Post.Servers[0].URL)

}