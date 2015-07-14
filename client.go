package osin

// Client information
type Client interface {
	// Client id
	GetId() string

	// Client secret
	GetSecret() string

	// Base client uri
	GetRedirectUri() string

	// Data to be passed to storage. Not used by the library.
	GetUserData() interface{}
}

// DefaultClient stores all data in struct variables
type DefaultClient struct {
	Id          string      `bson:"_id" json:"_id"` // how to fix this!!!!
	Secret      string      `bson:"secret" json:"secret"`
	RedirectUri string      `bson:"redirectUri" json:"redirectUri"`
	UserData    interface{} `bson:"userData" json:"userData"`
}

func (d DefaultClient) GetId() string {
	return d.Id
}

func (d DefaultClient) GetSecret() string {
	return d.Secret
}

func (d DefaultClient) GetRedirectUri() string {
	return d.RedirectUri
}

func (d DefaultClient) GetUserData() interface{} {
	return d.UserData
}

func (d DefaultClient) CopyFrom(client Client) {
	d.Id = client.GetId()
	d.Secret = client.GetSecret()
	d.RedirectUri = client.GetRedirectUri()
	d.UserData = client.GetUserData()
}
