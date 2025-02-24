package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io"
    "net/http"
)

func (kin *NomiKin) SendKindroidApiCall(endpoint string, method string, body interface{}, extraHeaders map[string]string) ([]byte, error) {
    headers := map[string]string{
        "Authorization": "Bearer " + kin.ApiKey,
        "Content-Type": "application/json",
    }
    
    if endpoint == "discord-bot" {
        if extraHeaders["X-Kindroid-Requester"] == "" {
            return nil, fmt.Errorf("Error: X-Kindroid-Requester header is required for discord-bot endpoint")
        }
        headers["X-Kindroid-Requester"] = extraHeaders["X-Kindroid-Requester"]
    }

    var jsonBody []byte
    var bodyReader io.Reader
    var err error

    if body != nil {
        jsonBody, err = json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("Error constructing body: %v: ", err)
        }
        bodyReader = bytes.NewBuffer(jsonBody)
    } else {
        bodyReader = nil
    }

    req, err := http.NewRequest(method, endpoint, bodyReader)
    if err != nil {
       return nil, fmt.Errorf("Error reading HTTP request: %v", err)
    }

    req.Header.Set("Authorization", headers["Authorization"])
    req.Header.Set("Content-Type", headers["Content-Type"])

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Error sending HTTP request: %v", err)
    }

    defer resp.Body.Close()

    bodyResponse, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Error response from Kin API: %v", string(bodyResponse))
    }

    log.Printf("Received reply message from Kin %v: %v", kin.CompanionId, string(bodyResponse))
    return bodyResponse, nil
}

func (kin *NomiKin) SendKindroidMessage(message *KinMessage) (*string, error) {
    log.Printf("Sending message to Kin %v: %v", kin.CompanionId, message.message)
    endpoint := UrlComponents["SendMessage"][0]
    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", message, nil)
    if err != nil {
        return nil, fmt.Errorf("Error sending message to Kin: %v", err)
    }


    kinReply := string(bodyResponse)
    return &kinReply, nil
}

func (kin *NomiKin) SendKindroidChatBreak(message *KinChatBreak) (*string, error) {
    log.Printf("Sending chat break to Kin %v: %v", kin.CompanionId, message.greeting)
    endpoint := UrlComponents["ChatBreak"][0]
    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", message, nil)
    if err != nil {
        return nil, fmt.Errorf("Error sending chat break to Kin: %v", err)
    }

    kinReply := string(bodyResponse)
    if kinReply == "" {
        kinReply = "Chat break successful"
    }
    return &kinReply, nil
}

func (kin *NomiKin) SendKindroidDiscordBot(kinShareId *string, discordNsfwFilter *bool, requester *string, conversation []KinConversation) (*string, error) {
    log.Printf("Sending message to Kin %v: %v messages", kin.CompanionId, len(conversation))
    endpoint := UrlComponents["DiscordBot"][0]
    extraHeaders := map[string]string{
        "X-Kindroid-Requester": *requester,
    }
    body := KinDiscordBot{
        share_code: *kinShareId,
        enable_filter: *discordNsfwFilter,
        conversation: conversation,
    }

    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", body, extraHeaders)
    if err != nil {
        return nil, fmt.Errorf("Error sending message to Kin: %v", err)
    }

    kinReply := string(bodyResponse)
    return &kinReply, nil
}
