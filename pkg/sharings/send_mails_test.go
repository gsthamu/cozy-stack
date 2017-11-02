package sharings

import (
	"fmt"
	"testing"

	"github.com/cozy/cozy-stack/client/auth"
	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/contacts"
	"github.com/cozy/cozy-stack/pkg/couchdb"
	"github.com/cozy/cozy-stack/pkg/permissions"
	"github.com/stretchr/testify/assert"
)

var rec = &contacts.Contact{
	Cozy:  []contacts.Cozy{},
	Email: []contacts.Email{},
}

var recStatus = &RecipientStatus{
	RefRecipient: couchdb.DocReference{
		Type: consts.Contacts,
	},
	Client: auth.Client{
		ClientID:     "",
		RedirectURIs: []string{},
	},
	recipient: rec,
}

var mailValues = &mailTemplateValues{}

var sharingTest = &Sharing{
	AppSlug:          "spartapp",
	SharingType:      consts.OneShotSharing,
	RecipientsStatus: []*RecipientStatus{recStatus},
	SharingID:        "sparta-id",
	Permissions:      permissions.Set{},
}

var instanceScheme = "http"

func TestGenerateMailMessageWhenRecipientHasNoEmail(t *testing.T) {
	msg, err := generateMailMessage(sharingTest, rec, mailValues)
	assert.Equal(t, ErrRecipientHasNoEmail, err)
	assert.Nil(t, msg)
}

func TestGenerateMailMessageSuccess(t *testing.T) {
	rec.Email = append(rec.Email, contacts.Email{Address: "this@mail.com"})
	_, err := generateMailMessage(sharingTest, rec, mailValues)
	assert.NoError(t, err)
}

func TestGenerateOAuthQueryStringWhenThereIsNoOAuthClient(t *testing.T) {
	// Without client id.
	recStatus.Client.RedirectURIs = []string{"redirect.me.to.heaven"}
	oauthQueryString, err := GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.Equal(t, ErrNoOAuthClient, err)
	assert.Equal(t, oauthQueryString, "")

	// Without redirect uri.
	recStatus.Client.ClientID = "sparta"
	recStatus.Client.RedirectURIs = []string{}
	oauthQueryString, err = GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.Equal(t, ErrNoOAuthClient, err)
	assert.Equal(t, oauthQueryString, "")

}

func TestGenerateOAuthQueryStringWhenRecipientHasNoURL(t *testing.T) {
	recStatus.Client.RedirectURIs = []string{"redirect.me.to.sparta"}

	oauthQueryString, err := GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.Equal(t, ErrRecipientHasNoURL, err)
	assert.Equal(t, "", oauthQueryString)
}

func TestGenerateOAuthQueryStringSuccess(t *testing.T) {
	rec.Cozy = make([]contacts.Cozy, 1)

	// First test: no scheme in the url.
	rec.Cozy[0].URL = "this.is.url"
	expectedStr := "http://this.is.url/sharings/request?App_slug=spartapp&client_id=sparta&redirect_uri=redirect.me.to.sparta&response_type=code&scope=&sharing_type=one-shot&state=sparta-id"

	oAuthQueryString, err := GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.NoError(t, err)
	assert.Equal(t, expectedStr, oAuthQueryString)

	// Second test: "http" scheme in the url.
	rec.Cozy[0].URL = "http://this.is.url"
	oAuthQueryString, err = GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.NoError(t, err)
	assert.Equal(t, expectedStr, oAuthQueryString)

	// Third test: "https" scheme in the url.
	rec.Cozy[0].URL = "https://this.is.url"
	expectedStr = "https://this.is.url/sharings/request?App_slug=spartapp&client_id=sparta&redirect_uri=redirect.me.to.sparta&response_type=code&scope=&sharing_type=one-shot&state=sparta-id"
	oAuthQueryString, err = GenerateOAuthQueryString(sharingTest, recStatus,
		instanceScheme)
	assert.NoError(t, err)
	assert.Equal(t, expectedStr, oAuthQueryString)
}

func TestSendSharingMails(t *testing.T) {
	rec.Cozy = make([]contacts.Cozy, 1)

	// Add the recipient in the database.
	rec.Cozy[0].URL = "this.is.url"
	rec.Email = nil
	err := couchdb.CreateDoc(in, rec)
	if err != nil {
		fmt.Printf("%v\n", err)
		t.Fail()
	}
	defer couchdb.DeleteDoc(in, rec)
	// Set the id to the id generated by Couch.
	recStatus.RefRecipient.ID = rec.DocID

	err = SendSharingMails(in, sharingTest)
	assert.Error(t, err)
}

func TestGenerateDiscoveryLinkRecipientHasNoEmail(t *testing.T) {
	recStatus.recipient.Email = nil
	_, err := generateDiscoveryLink(in, sharingTest, recStatus)
	assert.Equal(t, ErrRecipientHasNoEmail, err)
}

func TestGenerateDiscoveryLinkSuccess(t *testing.T) {
	email := contacts.Email{Address: "email.test"}
	recStatus.recipient.Email = append(recStatus.recipient.Email, email)
	expectedStr := "https://" + in.Domain + "/sharings/discovery?recipient_email=email.test&recipient_id=" + rec.ID() + "&sharing_id=sparta-id"

	discLink, err := generateDiscoveryLink(in, sharingTest, recStatus)
	assert.NoError(t, err)
	assert.Equal(t, expectedStr, discLink)
}
