package jamf

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

//Client lets you execute commands via the Jamf Pro API
type Client struct {
	BaseURL    string
	usr        string
	pw         string
	token      string
	httpClient http.Client
}

//GetTokenURL grabs the Token which lets you perform API Requests
func (j *Client) GetTokenURL() (string, error) {

	URL := fmt.Sprintf("%s/uapi/auth/tokens", j.BaseURL)

	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(j.usr, j.pw)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type tokenBody struct {
		Token   string
		Expires int
	}

	var tokenResponse tokenBody

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		panic(err)
	}

	return tokenResponse.Token, nil
}

//XMLResponseMobileDeviceGroups is the XML returned when asking for the Devices in the requested Group
type XMLResponseMobileDeviceGroups struct {
	XMLName       xml.Name      `xml:"mobile_device_group"`
	ID            int           `xml:"id"`
	MobileDevices MobileDevices `xml:"mobile_devices"`
}

//MobileDevices Struct found inside XMLResponseMobileDevicesGroups
type MobileDevices struct {
	Size         int            `xml:"size"`
	MobileDevice []MobileDevice `xml:"mobile_device"`
}

//MobileDevice contains the ID and Name of the Devices in the Group
type MobileDevice struct {
	ID   int    `xml:"id"`
	Name string `xml:"name"`
}

//GetDevicesInGroup takes in a group ID as a string
func (j *Client) GetDevicesInGroup(g string) ([]MobileDevice, error) {

	URL := fmt.Sprintf("%s/JSSResource/mobiledevicegroups/id/%s", j.BaseURL, g)

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(j.usr, j.pw)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v XMLResponseMobileDeviceGroups

	xml.Unmarshal(body, &v)

	return v.MobileDevices.MobileDevice, nil
}

//RestartDevice takes in the ID of a device that will be restarted
func (j *Client) RestartDevice(d int) (int, error) {

	URL := fmt.Sprintf("%s/JSSResource/mobiledevicecommands/command/RestartDevice/id/%d", j.BaseURL, d)

	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		return 500, err
	}

	req.SetBasicAuth(j.usr, j.pw)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return 500, err
	}

	return resp.StatusCode, nil
}

//RestartDeviceChannel takes in a channel of Mobile Device and returns a channel of results which includes MobileDevice plus a Success indicator
func (j *Client) RestartDeviceChannel(d MobileDevice, c chan string) {

	URL := fmt.Sprintf("%s/JSSResource/mobiledevicecommands/command/RestartDevice/id/%d", j.BaseURL, d.ID)

	success := false

	res := fmt.Sprintf("%d - %s restart: %t", d.ID, d.Name, success)

	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		c <- res
		return
	}

	req.SetBasicAuth(j.usr, j.pw)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		c <- res
		return
	}

	success = true

	res = fmt.Sprintf("%d - %s restart: %t, status code: %d", d.ID, d.Name, success, resp.StatusCode)

	c <- res

	return

}

//NewJamfClient requires BasicAuth to authenticate with JAMF Server
func NewJamfClient(URL string, username string, password string) *Client {

	return &Client{
		URL,
		username,
		password,
		"",
		http.Client{},
	}
}
