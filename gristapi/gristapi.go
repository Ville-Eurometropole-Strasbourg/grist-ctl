// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Grist API operation
package gristapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Grist's user
type User struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Access       string `json:"access"`
	ParentAccess string `json:"parentAccess"`
}

// Grist's Organization
type Org struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"createdAt"`
}

// Grist's workspace
type Workspace struct {
	Id                 int    `json:"id"`
	Name               string `json:"name"`
	CreatedAt          string `json:"createdAt"`
	Docs               []Doc  `json:"docs"`
	IsSupportWorkspace string `json:"isSupportWorkspace"`
	OrgDomain          string `json:"orgDomain"`
	Org                Org    `json:"org"`
	Access             string `json:"access"`
}

type EntityAccess struct {
	MaxInheritedRole string `json:"maxInheritedRole"`
	Users            []User `json:"users"`
}

// Grist's document
type Doc struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	IsPinned  bool      `json:"isPinned"`
	Workspace Workspace `json:"workspace"`
}

// Grist's table
type Table struct {
	Id string `json:"id"`
}

// List of Grist's tables
type Tables struct {
	Tables []Table `json:"tables"`
}

// Grist's table column
type TableColumn struct {
	Id string `json:"id"`
}

// List of Grist's table columns
type TableColumns struct {
	Columns []TableColumn `json:"columns"`
}

// Grist's table row
type TableRows struct {
	Id []uint `json:"id"`
}

// Grist's user role
type UserRole struct {
	Email string
	Role  string
}

// Apply config and return the config file path
func GetConfig() string {
	home := os.Getenv("HOME")
	configFile := filepath.Join(home, ".gristctl")
	if os.Getenv("GRIST_TOKEN") == "" || os.Getenv("GRIST_URL") == "" {
		err := godotenv.Load(configFile)
		if err != nil {
			fmt.Printf("Error reading configuration file : %s\n", err)
		}
	}
	return configFile
}

func init() {
	GetConfig()
}

// Sending an HTTP request to Grist's REST API
// Action: GET, POST, PATCH, DELETE
// Returns response body
func httpRequest(action string, myRequest string, data *bytes.Buffer) (string, int) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/%s", os.Getenv("GRIST_URL"), myRequest)
	bearer := "Bearer " + os.Getenv("GRIST_TOKEN")

	req, err := http.NewRequest(action, url, data)
	if err != nil {
		log.Fatalf("Error creating request %s: %s", url, err)
	}
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Error sending request %s: %s", url, err)
		return errMsg, -10
	} else {
		defer resp.Body.Close()
		// Read the HTTP response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response %s: %s", url, err)
		}
		return string(body), resp.StatusCode
	}
}

// Send an HTTP GET request to Grist's REST API
// Returns the response body
func httpGet(myRequest string, data string) (string, int) {
	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("GET", myRequest, dataBody)
	// if status != http.StatusOK {
	// 	fmt.Printf("Return code from %s : %d (%s)\n", myRequest, status, body)
	// }
	return body, status
}

// Test Grist API connection
func TestConnection() bool {
	_, status := httpGet("orgs", "")
	return status == http.StatusOK
}

// Sends an HTTP POST request to Grist's REST API with a data load
// Return the response body
func httpPost(myRequest string, data string) (string, int) {
	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("POST", myRequest, dataBody)
	return body, status
}

// Sends an HTTP PATCH request to Grist's REST API with a data load
// Return the response body
func httpPatch(myRequest string, data string) (string, int) {
	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("PATCH", myRequest, dataBody)
	return body, status
}

// Send an HTTP DELETE request to Grist's REST API with a data load
// Return the response body
func httpDelete(myRequest string, data string) (string, int) {
	dataBody := bytes.NewBuffer([]byte(data))
	body, status := httpRequest("DELETE", myRequest, dataBody)
	return body, status
}

// Retrieves the list of organizations
func GetOrgs() []Org {
	myOrgs := []Org{}
	response, _ := httpGet("orgs", "")
	json.Unmarshal([]byte(response), &myOrgs)
	return myOrgs
}

// Retrieves the organization whose identifier is passed in parameter
func GetOrg(idOrg string) Org {
	myOrg := Org{}
	response, _ := httpGet("orgs/"+idOrg, "")
	json.Unmarshal([]byte(response), &myOrg)
	return myOrg
}

// Retrieves the list of users in the organization whose ID is passed in parameter
func GetOrgAccess(idOrg string) []User {
	var lstUsers EntityAccess
	url := fmt.Sprintf("orgs/%s/access", idOrg)
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers.Users
}

// Retrieves information on a specific organization
func GetOrgWorkspaces(orgId int) []Workspace {
	lstWorkspaces := []Workspace{}
	response, _ := httpGet("orgs/"+strconv.Itoa(orgId)+"/workspaces", "")
	json.Unmarshal([]byte(response), &lstWorkspaces)
	return lstWorkspaces
}

// Get a workspace
func GetWorkspace(workspaceId int) Workspace {
	workspace := Workspace{}
	url := fmt.Sprintf("workspaces/%d", workspaceId)
	response, returnCode := httpGet(url, "")
	if returnCode == http.StatusOK {
		json.Unmarshal([]byte(response), &workspace)
	}
	return workspace
}

// Delete a workspace
func DeleteWorkspace(workspaceId int) {
	url := fmt.Sprintf("workspaces/%d", workspaceId)
	response, status := httpDelete(url, "")
	if status == http.StatusOK {
		fmt.Printf("Workspace %d deleted\t✅\n", workspaceId)
	} else {
		fmt.Printf("Unable to delete workspace %d : %s ❗️\n", workspaceId, response)
	}
}

// Delete a document
func DeleteDoc(docId string) {
	url := fmt.Sprintf("docs/%s", docId)
	response, status := httpDelete(url, "")
	if status == http.StatusOK {
		fmt.Printf("Document %s deleted\t✅\n", docId)
	} else {
		fmt.Printf("Unable to delete document %s : %s ❗️", docId, response)
	}
}

// Delete a user
func DeleteUser(userId int) {
	url := fmt.Sprintf("users/%d", userId)
	response, status := httpDelete(url, `{"name": ""}`)

	var message string
	switch status {
	case 200:
		message = "The account has been deleted successfully"
	case 400:
		message = "The passed user name does not match the one retrieved from the database given the passed user id"
	case 403:
		message = "The caller is not allowed to delete this account"
	case 404:
		message = "The user is not found"
	}
	fmt.Println(message)
	if status != http.StatusOK {
		fmt.Printf("ERREUR: %s\n", response)
	}
}

// Workspace access rights query
func GetWorkspaceAccess(workspaceId int) EntityAccess {
	workspaceAccess := EntityAccess{}
	url := fmt.Sprintf("workspaces/%d/access", workspaceId)
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &workspaceAccess)
	return workspaceAccess
}

// Retrieves information about a specific document
func GetDoc(docId string) Doc {
	doc := Doc{}
	url := "docs/" + docId
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &doc)
	return doc
}

// Retrieves the list of tables contained in a document
func GetDocTables(docId string) Tables {
	tables := Tables{}
	url := "docs/" + docId + "/tables"
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &tables)

	return tables
}

// Retrieves a list of table columns
func GetTableColumns(docId string, tableId string) TableColumns {
	columns := TableColumns{}
	url := "docs/" + docId + "/tables/" + tableId + "/columns"
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &columns)

	return columns
}

// Retrieves records from a table
func GetTableRows(docId string, tableId string) TableRows {
	rows := TableRows{}
	url := "docs/" + docId + "/tables/" + tableId + "/data"
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &rows)

	return rows
}

// Returns the list of users with access to the document
func GetDocAccess(docId string) EntityAccess {
	var lstUsers EntityAccess
	url := fmt.Sprintf("docs/%s/access", docId)
	response, _ := httpGet(url, "")
	json.Unmarshal([]byte(response), &lstUsers)
	return lstUsers
}

// Move all documents from a workspace to another
func MoveAllDocs(fromWorkspaceId int, toWorkspaceId int) {
	// Getting the workspaces
	from_ws := GetWorkspace(fromWorkspaceId)
	to_ws := GetWorkspace(toWorkspaceId)
	if from_ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", fromWorkspaceId)
	} else if to_ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", toWorkspaceId)
	} else {
		// Workspaces were found
		for _, doc := range from_ws.Docs {
			url := "docs/" + doc.Id + "/move"
			data := fmt.Sprintf(`{"workspace": "%d"}`, toWorkspaceId)
			_, status := httpPatch(url, data)
			if status == http.StatusOK {
				fmt.Printf("Document %s moved to workspace %d ✅\n", doc.Id, toWorkspaceId)
			} else {
				fmt.Printf("Unable to move document %s", doc.Id)
			}
		}
	}
}

// Move a document in a workspace
func MoveDoc(docId string, workspaceId int) {
	url := "docs/" + docId + "/move"
	data := fmt.Sprintf(`{"workspace": "%d"}`, workspaceId)
	_, status := httpPatch(url, data)
	if status == http.StatusOK {
		fmt.Printf("Document moved to workspace %d ✅\n", workspaceId)
	} else {
		fmt.Printf("Unable to move document")
	}
}

// Purge a document's history, to retain only the last modifications
func PurgeDoc(docId string, nbHisto int) {
	url := "docs/" + docId + "/states/remove"
	data := fmt.Sprintf(`{"keep": "%d"}`, nbHisto)
	_, status := httpPost(url, data)
	if status == http.StatusOK {
		fmt.Printf("History cleared (%d last states) ✅\n", nbHisto)
	}
}

// Import a list of user & role into a workspace
// Search workspace by name in org
func ImportUsers(orgId int, workspaceName string, users []UserRole) {
	lstWorkspaces := GetOrgWorkspaces(orgId)
	idWorkspace := 0
	for _, ws := range lstWorkspaces {
		if ws.Name == workspaceName {
			idWorkspace = ws.Id
		}
	}

	if idWorkspace == 0 {
		idWorkspace = CreateWorkspace(orgId, workspaceName)
	}
	if idWorkspace == 0 {
		fmt.Printf("Unable to create workspace %s\n", workspaceName)
	} else {
		url := fmt.Sprintf("workspaces/%d/access", idWorkspace)

		roleLine := []string{}
		for _, role := range users {
			roleLine = append(roleLine, fmt.Sprintf(`"%s": "%s"`, role.Email, role.Role))
		}
		patch := fmt.Sprintf(`{	"delta": { "users": {%s}}}`, strings.Join(roleLine, ","))

		body, status := httpPatch(url, patch)

		var result string
		if status == http.StatusOK {
			result = "✅"
		} else {
			result = fmt.Sprintf("❗️ (%s)", body)
		}
		fmt.Printf("Import %d users in workspace n°%d\t : %s\n", len(users), idWorkspace, result)
	}

}

// Create a workspace in an organization
func CreateWorkspace(orgId int, workspaceName string) int {
	url := fmt.Sprintf("orgs/%d/workspaces", orgId)
	data := fmt.Sprintf(`{"name":"%s"}`, workspaceName)
	body, status := httpPost(url, data)
	idWorkspace := 0
	if status == http.StatusOK {
		id, err := strconv.Atoi(body)
		if err == nil {
			idWorkspace = id
		}
	}
	return idWorkspace
}

// Export doc in Grist format (Sqlite) in fileName file
func ExportDocGrist(docId string, fileName string) {
	url := fmt.Sprintf("docs/%s/download", docId)
	export, returnCode := httpGet(url, "")
	if returnCode == http.StatusOK {
		f, e := os.Create(fileName)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		fmt.Fprintln(f, export)
	}
}

// Export doc in Excel format (XLSX) in fileName file
func ExportDocExcel(docId string, fileName string) {
	url := fmt.Sprintf("docs/%s/download/xlsx", docId)
	export, returnCode := httpGet(url, "")
	if returnCode == http.StatusOK {
		f, e := os.Create(fileName)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		fmt.Fprintln(f, export)
	}
}

// Returns table content as Dataframe
func GetTableContent(docId string, tableName string) {
	url := fmt.Sprintf("docs/%s/download/csv?tableId=%s", docId, tableName)
	csvFile, _ := httpGet(url, "")
	fmt.Println(csvFile)
}
