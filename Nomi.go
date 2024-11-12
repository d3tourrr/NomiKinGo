package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io"
    "io/ioutil"
    "net/http"
)

var UrlComponents map[string][]string

type Room struct {
    Name string
    Note string
    backchannelingEnabled bool
    Nomis []string
    Uuid string
    Status string
}

func (nomi *NomiKin) Init() {
    log.Println("Entered Init")
    UrlComponents = make(map[string][]string)
    UrlComponents["SendMessage"] = []string {"https://api.nomi.ai/v1/nomis", "chat"}
    UrlComponents["RoomCreate"] = []string {"https://api.nomi.ai/v1/rooms"}
    UrlComponents["RoomSend"] = []string {"https://api.nomi.ai/v1/rooms", "chat"}
    UrlComponents["RoomReply"] = []string {"https://api.nomi.ai/v1/rooms", "chat/request"}
}

func (nomi *NomiKin) ApiCall(endpoint string, method string, body interface{}) ([]byte, error) {
    headers := map[string]string{
        "authorization": nomi.ApiKey,
        "content-type": "application/json",
    }

    var jsonBody []byte
    var bodyReader io.Reader
    var err error
    if body != nil {
        jsonBody, err = json.Marshal(body)
        bodyReader = bytes.NewBuffer(jsonBody)
        if err != nil {
            return nil, fmt.Errorf("Error constructing body: %v: ", err)
        }
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
        return nil, fmt.Errorf("Error making HTTP request: %v", err)
    }

    defer resp.Body.Close()

    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        var errorResult map[string]interface{}
        if err := json.Unmarshal(responseBody, &errorResult); err != nil {
            return nil, fmt.Errorf("Error unmarshalling error response: %v\n%v", err, string(responseBody))
        }

        return nil, fmt.Errorf("Error response from Nomi API: %v", string(responseBody))
    }

    return responseBody, nil
}

func (nomi *NomiKin) RoomExists(roomName *string) (bool, error) {
    log.Printf("Checking Nomi %v room %v", nomi.CompanionId, *roomName)
    roomUrl := UrlComponents["RoomCreate"][0]
    roomResult, err := nomi.ApiCall(roomUrl, "Get", nil)
    var rooms []Room

    if err != nil {
        return false, err
    }

    if err := json.Unmarshal([]byte(roomResult), &rooms); err != nil {
        return false, err
    }

    exists := false
    for _, r := range rooms {
        log.Printf("Checking room %v - %v", r.Name, r.Uuid)
        if r.Name == *roomName {
            exists = true
            break
        }
    }

    return exists, nil
}

func (nomi *NomiKin) CreateNomiRoom(name *string, note *string, backchannelingEnabled *bool, nomiUuids []string) (string, error) {
    roomExists, err := nomi.RoomExists(name)
    if err != nil {
        log.Printf("Error checking if room exists: %v", err)
    }

    log.Printf("Room exists: %v", roomExists)
    if roomExists {
        // TODO: Add the Nomi to the room
        return *name, nil
    }

    roomUrl := UrlComponents["RoomCreate"][0]
    bodyMap := map[string]interface{}{
        "name": *name,
        "note": *note,
        "backchannelingEnabled": backchannelingEnabled,
        "nomiUuids": nomiUuids,
    }

    response, err := nomi.ApiCall(roomUrl, "Post", bodyMap)
    if err != nil {
        return "", err
    }

    var result map[string]interface{}
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        if roomCreateName, ok := result["name"].(string); ok {
            log.Printf("Created Nomi %v room: %v\n", nomi.CompanionId, roomCreateName)
            return roomCreateName, nil
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}

func (nomi *NomiKin) SendNomiRoomMessage(message *string, roomId *string) (string, error) {
    if len(*message) > 599 {
        log.Printf("Message too long: %d", len(*message))
        return fmt.Sprintf("Your message was `%d` characters long, but the maximum message length is 600. Please send a shorter message.", len(*message)), nil
    }

    bodyMap := map[string]string{
        "messageText": *message,
    }

    messageSendUrl := UrlComponents["RoomSend"][0] + "/" + *roomId + "/" + UrlComponents["RoomSend"][1]
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }

    var result map[string]interface{}
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        if replyMessage, ok := result["sentMessage"].(map[string]interface{}); ok {
            log.Printf("Sent message to room %v: %v\n", roomId, replyMessage)
            return fmt.Sprintf("Sent message to room %v: %v\n", roomId, replyMessage), nil
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}

func (nomi *NomiKin) RequestNomiRoomReply(roomId *string, nomiId *string) (string, error) {
    bodyMap := map[string]string{
        "nomiUuid": *nomiId,
    }

    messageSendUrl := UrlComponents["RoomReply"][0] + "/" + *roomId + "/" + UrlComponents["RoomReply"][1]
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }


    var result map[string]interface{}
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        if replyMessage, ok := result["sentMessage"].(map[string]interface{}); ok {
            log.Printf("Sent message to room %v: %v\n", roomId, replyMessage)
            return fmt.Sprintf("Sent message to room %v: %v\n", roomId, replyMessage), nil
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}

func (nomi *NomiKin) SendNomiMessage (message *string) (string, error) {
    if len(*message) > 599 {
        log.Printf("Message too long: %d", len(*message))
        return fmt.Sprintf("Your message was `%d` characters long, but the maximum message length is 600. Please send a shorter message.", len(*message)), nil
    }

    bodyMap := map[string]string{
        "messageText": *message,
    }

    messageSendUrl := UrlComponents["SendMessage"][0] + "/" + nomi.CompanionId + "/chat"
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }

    var result map[string]interface{}
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        if replyMessage, ok := result["replyMessage"].(map[string]interface{}); ok {
            log.Printf("Received reply message from Nomi %v: %v\n", nomi.CompanionId, replyMessage)
            if textValue, ok := replyMessage["text"].(string); ok {
                return textValue, nil
            }
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}
