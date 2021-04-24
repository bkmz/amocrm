package models

import (
	"time"
	"fmt"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type (
	Ct struct {
		request request
	}

	// TagsField struct {
	// 	Id   int    `json:"id"`
	// 	Name string `json:"name"`
	// }

	contact struct {
		Id                int
		Name              string
		ResponsibleUserId int   `json:"responsible_user_id"`
		CreatedBy         int   `json:"created_by"`
		CreatedAt         int64 `json:"created_at"`
		UpdatedAt         int64 `json:"updated_at"`
		AccountId         int   `json:"account_id"`
		UpdatedBy         int   `json:"updated_by"`
		GroupId           int   `json:"group_id"`
		Company           struct {
			Id   int
			Name string
		}
		Leads struct {
			Id []int `json:"id"`
		} `json:"leads"`
		Tags          []TagsField   `json:"tags"`
		ClosestTaskAt int           `json:"closest_task_at"`
		CustomFields  []CustomField `json:"custom_fields"`
	}

	// TODO: Need add part of stuct to unmarshal all resp
	// {"_links":{"self":{"href":"\/api\/v2\/contacts","method":"post"}},"_embedded":{"items":[],"errors":{"update":{"59144296":"Last modified date is older than in database"}}}}

	// {"_links":{"self":{"href":"\/api\/v2\/contacts","method":"post"}},"_embedded":{"items":[{"id":59144296,"updated_at":1619342253000,"_links":{"self":{"href":"\/api\/v2\/contacts?id=59144296","method":"get"}}}]}}

	allContacts struct {
		Links struct {
			Self struct {
				Href   string `json:"href"`
				Method string `json:"method"`
			} `json:"self"`
		} `json:"_links"`
		Embedded struct {
			Items []*contact
			Errors map[string]map[int]string `json:"errors"`
		} `json:"_embedded"`
	}
)

// Method creates empty struct
func (c Ct) Create() *contact {
	return &contact{}
}

// Method gets all contacts from API AmoCRM
//
// Example
//    api := amocrm.NewAmo("login", "key", "domain")
//    allLeads, _ := api.Contact.All()
func (c Ct) All() ([]*contact, error) {
	return c.multiplyRequest(contactUrl)
}

// Method gets all contacts by responsible from API AmoCRM
//
// Example
//    api := amocrm.NewAmo("login", "key", "domain")
//    leads, _ := api.Lead.Responsible(12345)
func (c Ct) Responsible(id int) ([]*contact, error) {
	url := constructUrlWithResponsible(contactUrl, id)
	return c.multiplyRequest(url)
}

// Method gets all contacts by query from API AmoCRM
//
// Example
//    api := amocrm.NewAmo("login", "key", "domain")
//    leads, _ := api.Contact.Query("+79671234567")
func (c Ct) Query(query string) ([]*contact, error) {
	url := constructUrlWithQuery(contactUrl, query)
	return c.multiplyRequest(url)
}

func (c Ct) multiplyRequest(url string) ([]*contact, error) {
	contacts := allContacts{}
	// API returns only 500 rows per request
	// this loop count necessary offset and request data again
	for i := 0; ; i++ {
		var tmpContacts allContacts
		resultJson, err := c.request.get(constructUrlWithOffset(url, i))
		if err != nil {
			return nil, err
		}
		json.Unmarshal(resultJson, &tmpContacts)
		// sets current data after request to general slice
		contacts.Embedded.Items = append(contacts.Embedded.Items, tmpContacts.Embedded.Items...)
		if len(tmpContacts.Embedded.Items) < 500 {
			break
		}
	}

	return contacts.Embedded.Items, nil
}

// Method gets only one row by ID
func (c Ct) Id(id int) (*contact, error) {
	var contacts allContacts
	url := constructUrlWithId(contactUrl, id)
	resultJson, err := c.request.get(url)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(resultJson, &contacts)
	return contacts.Embedded.Items[0], nil
}

// Note:
//    Field `Name` is required
//    Return id of new contact
//
// Example:
//    api := amocrm.NewAmo("login", "key", "domain")
//    contact := api.Contact.Create()
//    contact.Name = "test"
//    id, err := api.Contact.Add(lead)
func (c Ct) Add(ct *contact) (int, error) {
	data := map[string]interface{}{}
	data["name"] = ct.Name
	if ct.ResponsibleUserId != 0 {
		data["responsible_user_id"] = ct.ResponsibleUserId
	}
	if ct.Company.Id != 0 {
		data["company_id"] = ct.Company.Id
	}
	if len(ct.CustomFields) != 0 {
		data["custom_fields"] = ct.CustomFields
	}
	if len(ct.Tags) != 0 {
		res := make([]int, 0)
		for _, val := range ct.Tags {
			res = append(res, val.Id)
		}
		data["tags"] = res
	}
	fullData := map[string][]interface{}{"add": {data}}
	jsonData, _ := json.Marshal(fullData)

	log.WithFields(log.Fields{
		"data": fmt.Sprintf("%s", jsonData),
	}).Debug("Sending data")

	resp, err := c.request.Post(contactUrl, jsonData)
	if err != nil {
		return 0, err
	}

	log.WithFields(log.Fields{
		"data": fmt.Sprintf("%s", resp),
	}).Debug("Responce data")

	var newContact allContacts
	json.Unmarshal(resp, &newContact)
	return newContact.Embedded.Items[0].Id, nil
}

// Note:
//    Id is required
//
// Example:
//	   api := amocrm.NewAmo("login", "key", "domain")
//	   contact, _ := api.Contact.Id(123456)
//	   contact.Name = "test"
//	   _ = api.Contact.Update(contact)
func (c Ct) Update(ct *contact) error {
	data := map[string]interface{}{}
	data["id"] = ct.Id
	data["name"] = ct.Name
	// data["updated_at"] = ct.UpdatedAt + 1
	data["updated_at"] = time.Now().Unix()
	if ct.Company.Id != 0 {
		data["company_id"] = ct.Company.Id
	}
	data["responsible_user_id"] = ct.ResponsibleUserId
	data["custom_fields"] = ct.CustomFields
	data["created_by"] = ct.CreatedBy
	if len(ct.Tags) != 0 {
		res := make([]int, 0)
		for _, val := range ct.Tags {
			res = append(res, val.Id)
		}
		data["tags"] = res
	}

	fullData := map[string][]interface{}{"update": {data}}
	jsonData, _ := json.Marshal(fullData)
	log.WithFields(log.Fields{
		"data": fmt.Sprintf("%s", jsonData),
	}).Debug("Sending data")
	resp, err := c.request.Post(contactUrl, jsonData)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"data": fmt.Sprintf("%s", resp),
	}).Debug("Responce data")

	var newContact allContacts
	json.Unmarshal(resp, &newContact)
	if newContact.Embedded.Errors != nil {
		return fmt.Errorf("Can't update contact %d, reason: %s", ct.Id, newContact.Embedded.Errors["update"][ct.Id])
	}
	return nil
}
