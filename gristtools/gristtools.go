// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Common tools for Grist
package gristtools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gristctl/common"
	"gristctl/gristapi"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/go-gota/gota/dataframe"
	"github.com/olekukonko/tablewriter"
)

var output string

func SetOutput(out string) {
	output = out
}

// Display help message and quit
func Help() {

	common.DisplayTitle(common.T("app.title"))
	type command struct {
		cmd  string
		help string
	}

	commands := []command{
		{"config", common.T("help.config")},
		{"delete doc <id>", common.T("help.deleteDoc")},
		{"delete user <id>", common.T("help.deleteUser")},
		{"delete workspace <id>", common.T("help.deleteWorkspace")},
		{"[-o=json/table] get doc <id> access", common.T("help.docAccess")},
		{"get doc <id> excel", common.T("help.docExportExcel")},
		{"get doc <id> grist", common.T("help.docExportGrist")},
		{"get doc <id> table <tableName>", common.T("help.docExportCsv")},
		{"[-o=json/table] get doc <id>", common.T("help.docDesc")},
		{"[-o=json/table] get org <id>", common.T("help.orgDesc")},
		{"[-o=json/table] get org", common.T("help.orgList")},
		{"[-o=json/table] get users", common.T("help.userList")},
		{"[-o=json/table] get workspace <id> access", common.T("help.workspaceAccess")},
		{"[-o=json/table] get workspace <id>", common.T("help.workspaceDesc")},
		{"import users", common.T("help.userImport")},
		{"move doc <id> workspace <workspaceId>", common.T("help.docMove")},
		{"purge doc <id> [<number of states to keep>]", common.T("help.docPurge")},
		{"version", common.T("help.version")},
	}
	// Sort commands by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].cmd < commands[j].cmd
	})

	fmt.Printf("%s :\n", common.T("help.accepted"))
	for _, command := range commands {
		fmt.Print("- ")
		common.PrintCommand(command.cmd)
		fmt.Print(" : ")
		fmt.Println(command.help)
	}
	os.Exit(0)
}

// Displays the version of the program
func Version(version string) {
	fmt.Println("Version : ", version)
}

/*
Configure Grist envfile (url and api token)
Interactive filling the `.gristctl` file
*/
func Config() {
	configFile := gristapi.GetConfig()
	common.DisplayTitle(fmt.Sprintf("%s (%s)", common.T("config.title"), configFile))
	fmt.Printf("%s :\n- URL : %s\n", common.T("config.actual"), os.Getenv("GRIST_URL"))
	token := ""
	for i := 0; i < len(os.Getenv("GRIST_TOKEN")); i++ {
		token += "•"
	}
	fmt.Printf("- %s : %s\n", common.T("config.token"), token)
	testConnect := "❌"
	if gristapi.TestConnection() {
		testConnect = "✅"
	}
	fmt.Printf("%s : %s\n", common.T("config.connectTest"), testConnect)

	if common.Confirm(common.T("config.config")) {
		var url string
		urlSet := false
		for urlSet == false {
			url = common.Ask(common.T("config.urlSet"))

			// Test if url is well formatted
			urlOk, _ := regexp.MatchString(`^https?://.*[^/]$`, url)
			urlSet = urlOk
		}
		var token = common.Ask(common.T("config.token"))
		if common.Confirm(fmt.Sprintf("\n%s :\n- URL : %s\n- Token: %s\n%s ", common.T("config.new"), url, token, common.T("questions.isOk"))) {
			f, err := os.Create(configFile)
			if err != nil {
				fmt.Printf("%s %s (%s)", common.T("config.saveError"), configFile, err)
				os.Exit(-1)
			}
			defer f.Close()
			config := fmt.Sprintf("GRIST_URL=\"%s\"\nGRIST_TOKEN=\"%s\"\n", url, token)
			f.WriteString(config)
			f.Close()
			fmt.Printf("%s %s\n", common.T("config.savedIn"), configFile)

			// Test the configuration by connecting to the server
			nbOrgs := len(gristapi.GetOrgs())
			fmt.Printf("Nb orgs : %d\n", nbOrgs)
			if nbOrgs <= 0 {
				fmt.Println(common.T("config.connectError"))
				os.Exit(-1)
			}
		}
	}
}

/*
User role translation

Returns the role explanation corresponding to its code
*/
func TranslateRole(roleCode string) string {
	var role string
	switch roleCode {
	case "":
		role = "No inheritance of rights from upper level"
	case "owners":
		role = "Full inheritance of rights from the next level up"
	case "editors":
		role = "Inherit display and edit rights from higher level"
	case "viewers":
		role = "Inheritance of consultation rights from higher level"
	default:
		role = fmt.Sprintf("Inheritance level : %s\n", roleCode)
	}
	return role
}

/*
Import users from a list sent to standard input (stdin)

CSV input file has to be formatied with the following columns, separated with ';' :
- mail
- org id
- Workspace name
- role

Missing workspaces will be created on import.
*/
func ImportUsers() {
	common.DisplayTitle("Import users from stdin")
	fmt.Println("Expected data format : <mail>;<org id>;<workspace name>;<role>")

	scanner := bufio.NewScanner(os.Stdin)
	type userAccess struct {
		Mail          string
		OrgId         int
		WorkspaceName string
		Role          string
	}
	lstUserAccess := []userAccess{}
	for scanner.Scan() {
		line := scanner.Text()
		data := strings.Split(line, ";")
		if len(data) == 4 {
			var lineOk bool = true
			newUserAccess := userAccess{}
			newUserAccess.Mail = data[0]
			if !common.IsValidEmail(newUserAccess.Mail) {
				lineOk = false
			}
			orgId, errOrg := strconv.Atoi(data[1])
			if errOrg != nil {
				fmt.Printf("ERROR : org id should be an integer : %s\n", data[1])
				lineOk = false
			}
			newUserAccess.OrgId = orgId
			newUserAccess.WorkspaceName = data[2]
			newUserAccess.Role = data[3]
			if lineOk {
				lstUserAccess = append(lstUserAccess, newUserAccess)
			} else {
				fmt.Printf("ERROR : badly formatted line : %s\n", line)
			}
		} else {
			fmt.Printf("ERROR : badly formatted line (should have 4 columns): %s\n", line)
		}
	}

	if scanner.Err() != nil {
		fmt.Println("Standard input read error")
	}
	usersDf := dataframe.LoadStructs(lstUserAccess)

	workspaces := usersDf.GroupBy("OrgId", "WorkspaceName")
	for group, users := range workspaces.GetGroups() {
		var roles []gristapi.UserRole
		line := strings.Split(group, "_")
		orgId, orgErr := strconv.Atoi(line[0])
		if orgErr != nil {
			Help()
		}
		workspaceId := line[1]
		for i, user := range users.Select([]string{"Mail", "Role"}).Records() {
			if i > 0 {
				newRole := gristapi.UserRole{Email: user[0], Role: user[1]}
				roles = append(roles, newRole)
			}
		}
		gristapi.ImportUsers(orgId, workspaceId, roles)
	}
}

// Displays the list of users witch access to an organization
func DisplayOrgAccess(idOrg string) {

	lstUsers := gristapi.GetOrgAccess(idOrg)

	switch output {
	case "table":
		{
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Email", "Name", "Access"})
			for _, user := range lstUsers {
				table.Append([]string{user.Email, user.Name, user.Access})
			}

			table.Render()
		}
	case "json":
		{
			jsonUsers, err := json.MarshalIndent(lstUsers, "", "  ")
			if err != nil {
				fmt.Println("ERROR :", err)
			}
			fmt.Println(string(jsonUsers))
		}
	}
}

/*
	Displays detailed information about a document

- Document name
- Number of tables
- For each table :
  - Number of columns
  - Number of rows
  - List of columns
*/
func DisplayDoc(docId string) {
	type TableDetails struct {
		Name       string
		Nb_rows    int
		Nb_cols    int
		Cols_names []string
	}

	type DocInfo struct {
		Id       string         `json:"id"`
		Name     string         `json:"name"`
		IsPinned bool           `json:"pined"`
		NbTables int            `json:"nbTables"`
		Tables   []TableDetails `json:"tables"`
	}

	// Getting the document
	doc := gristapi.GetDoc(docId)
	if doc.Id == "" {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	} else {
		// Document was found
		// Getting the doc's tables
		var tables gristapi.Tables = gristapi.GetDocTables(docId)

		myDoc := DocInfo{
			Id:       doc.Id,
			Name:     doc.Name,
			IsPinned: doc.IsPinned,
			NbTables: len(tables.Tables),
		}

		// Getting the tables details
		var wg sync.WaitGroup
		var tables_details []TableDetails
		for _, table := range tables.Tables {
			wg.Add(1)
			go func() {
				defer wg.Done()
				table_desc := ""
				columns := gristapi.GetTableColumns(docId, table.Id)
				rows := gristapi.GetTableRows(docId, table.Id)

				var cols_names []string
				for _, col := range columns.Columns {
					cols_names = append(cols_names, col.Id)
				}
				slices.Sort(cols_names)
				for _, col := range cols_names {
					table_desc += fmt.Sprintf("%s ", col)
				}
				table_info := TableDetails{
					Name:       table.Id,
					Nb_rows:    len(rows.Id),
					Nb_cols:    len(columns.Columns),
					Cols_names: cols_names,
				}
				tables_details = append(tables_details, table_info)
			}()
		}
		wg.Wait()

		myDoc.Tables = tables_details

		switch output {
		case "json":
			{
				jsonDoc, err := json.MarshalIndent(myDoc, "", "   ")
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonDoc))
			}
		case "table":
			{
				// Displaying the document name
				pinned := ""
				if myDoc.IsPinned {
					pinned = "📌"
				}
				common.DisplayTitle(fmt.Sprintf("Document '%s' (%s) %s", myDoc.Name, myDoc.Id, pinned))
				fmt.Printf("Contains %d tables :\n", myDoc.NbTables)
				// Displaying the tables details
				tableView := tablewriter.NewWriter(os.Stdout)
				tableView.SetHeader([]string{"Table", common.T("col.nbCols"), common.T("col.columns"), common.T("col.nbRows")})
				for _, table_details := range tables_details {
					for i, col_name := range table_details.Cols_names {
						if i == 0 {
							tableView.Append([]string{table_details.Name, strconv.Itoa(table_details.Nb_cols), col_name, strconv.Itoa(table_details.Nb_rows)})
						} else {
							tableView.Append([]string{"", "", col_name, ""})
						}
					}
				}
				tableView.Render()
			}
		}
	}

}

// Displays the list of accessible organizations
func DisplayOrgs() {

	// Getting the list of organizations
	lstOrgs := gristapi.GetOrgs()
	// Sorting the list of organizations by name (lowercase)
	sort.Slice(lstOrgs, func(i, j int) bool {
		return strings.ToLower(lstOrgs[i].Name) < strings.ToLower(lstOrgs[j].Name)
	})
	table := tablewriter.NewWriter(os.Stdout)

	switch output {
	case "table":
		{
			table.SetHeader([]string{common.T("col.ident"), common.T("col.name")})
			for _, org := range lstOrgs {
				table.Append([]string{strconv.Itoa(org.Id), org.Name})
			}
			table.Render()
		}
	case "json":
		{
			jsonOrgs, err := json.MarshalIndent(lstOrgs, "", "  ")
			if err != nil {
				fmt.Println("ERROR :", err)
			}
			fmt.Println(string(jsonOrgs))
		}
	}
}

// Displays details about an organization
func DisplayOrg(orgId string) {

	type WpDesc struct {
		Id     int    `json:"id"`
		Name   string `json:"name"`
		NbDoc  int    `json:"nbDoc"`
		NbUser int    `json:"nbUser"`
	}

	type OrgDesc struct {
		Id   int      `json:"id"`
		Name string   `json:"name"`
		NbWs int      `json:"nbWs"`
		Ws   []WpDesc `json:"ws"`
	}

	var lstWsDesc []WpDesc

	org := gristapi.GetOrg(orgId)
	if org.Id == 0 {
		fmt.Printf("❗️ Organization %s not found ❗️\n", orgId)
	} else {

		// Org was found
		worskspaces := gristapi.GetOrgWorkspaces(org.Id)
		var wg sync.WaitGroup
		// Retrieving the number of documents and users for each workspace
		for _, ws := range worskspaces {
			func() {
				defer wg.Done()
				wg.Add(1)
				users := gristapi.GetWorkspaceAccess(ws.Id)
				nbUsers := 0
				for _, user := range users.Users {
					if user.Access != "" {
						nbUsers += 1
					}
				}
				lstWsDesc = append(lstWsDesc, WpDesc{ws.Id, ws.Name, len(ws.Docs), nbUsers})
			}()
		}
		wg.Wait()
		// Sorting the list of workspaces by name
		sort.Slice(lstWsDesc, func(i, j int) bool {
			return lstWsDesc[i].Name < lstWsDesc[j].Name
		})
		switch output {
		case "table":
			{
				common.DisplayTitle(fmt.Sprintf("%s n°%d : %s", common.T("org.name"), org.Id, org.Name))
				fmt.Printf("%s %d:\n", common.T("org.contains"), len(worskspaces))
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{common.T("col.ident"), common.T("col.name"), common.T("col.nbDocs"), common.T("col.directUsers")})
				// Displaying the list of workspaces
				for _, desc := range lstWsDesc {
					table.Append([]string{strconv.Itoa(desc.Id), desc.Name, strconv.Itoa(desc.NbDoc), strconv.Itoa(desc.NbUser)})
				}
				table.Render()
			}
		case "json":
			{
				myOrg := OrgDesc{
					Id:   org.Id,
					Name: org.Name,
					NbWs: len(worskspaces),
					Ws:   lstWsDesc,
				}

				jsonData, err := json.MarshalIndent(myOrg, "", "  ")
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonData))
			}
		}
	}
}

// Display a Workspace
func DisplayWorkspace(workspaceId int) {

	type docDesc struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		IsPinned bool   `json:"isPinned"`
	}

	type WorkspaceDesc struct {
		OrgId   int       `json:"orgId"`
		OrgName string    `json:"orgName"`
		Id      int       `json:"id"`
		Name    string    `json:"name"`
		NbDocs  int       `json:"nbDocs"`
		Docs    []docDesc `json:"docs"`
	}

	// Getting the workspace
	ws := gristapi.GetWorkspace(workspaceId)
	if ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", workspaceId)
	} else {
		// Workspace was found

		myDocs := []docDesc{}
		for _, doc := range ws.Docs {
			myDocs = append(myDocs, docDesc{doc.Id, doc.Name, doc.IsPinned})
		}

		// Sort the documents by name (lowercase)
		sort.Slice(myDocs, func(i, j int) bool {
			return strings.ToLower(myDocs[i].Name) < strings.ToLower(myDocs[j].Name)
		})

		myWS := WorkspaceDesc{
			OrgId:   ws.Org.Id,
			OrgName: ws.Org.Name,
			Id:      ws.Id,
			Name:    ws.Name,
			NbDocs:  len(ws.Docs),
			Docs:    myDocs,
		}

		switch output {
		case "table":
			{
				common.DisplayTitle(fmt.Sprintf("%s n°%d : '%s' | %s n°%d : '%s'",
					common.T("org.name"),
					myWS.OrgId,
					myWS.OrgName,
					common.T("workspace.name"),
					myWS.Id,
					myWS.Name))
				fmt.Printf("Contains %d documents :\n", myWS.NbDocs)
				// Listing the documents
				if myWS.NbDocs > 0 {
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{common.T("col.ident"), common.T("col.name"), common.T("col.pinned")})
					for _, doc := range myWS.Docs {
						pin := ""
						if doc.IsPinned {
							pin = "📌"
						}
						table.Append([]string{doc.Id, doc.Name, pin})
					}
					table.Render()
				} else {
					fmt.Println("No documents")
				}
			}
		case "json":
			{
				jsonData, err := json.MarshalIndent(myWS, "", "  ")
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonData))
			}
		}
	}
}

// Displays workspace access rights
func DisplayWorkspaceAccess(workspaceId int) {
	type wsUser struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Name         string `json:"name"`
		ParentAccess string `json:"parentAccess"`
		Access       string `json:"access"`
	}

	type wsAccess struct {
		WokspaceId       int      `json:"workspaceId"`
		WorkspaceName    string   `json:"workspaceName"`
		OrgId            int      `json:"orgId"`
		OrgName          string   `json:"orgName"`
		NbUsers          int      `json:"nbUsers"`
		MaxInheritedRole string   `json:"maxInheritedRole"`
		Users            []wsUser `json:"users"`
	}

	// Getting the workspace
	ws := gristapi.GetWorkspace((workspaceId))
	if ws.Id == 0 {
		fmt.Printf("❗️ Workspace %d not found ❗️\n", workspaceId)
	} else {
		// Workspace was found
		wsa := gristapi.GetWorkspaceAccess(workspaceId)

		var myUsers []wsUser
		nbUsers := 0
		for _, user := range wsa.Users {
			if user.Access != "" || user.ParentAccess != "" {
				tmpUser := wsUser{
					Id:           user.Id,
					Email:        user.Email,
					Name:         user.Name,
					ParentAccess: user.ParentAccess,
					Access:       user.Access,
				}
				myUsers = append(myUsers, tmpUser)
				nbUsers++
			}
		}
		// Sort users by email (lowercase)
		sort.Slice(myUsers, func(i, j int) bool {
			return strings.ToLower(myUsers[i].Email) < strings.ToLower(myUsers[j].Email)
		})
		myWsAccess := wsAccess{
			WokspaceId:       ws.Id,
			WorkspaceName:    ws.Name,
			OrgId:            ws.Org.Id,
			OrgName:          ws.Org.Name,
			MaxInheritedRole: wsa.MaxInheritedRole,
			NbUsers:          nbUsers,
			Users:            myUsers,
		}

		switch output {
		case "table":
			{
				// Displaying the workspace name
				common.DisplayTitle(fmt.Sprintf("Workspace n°%d : %s", myWsAccess.WokspaceId, myWsAccess.WorkspaceName))

				// Displaying the MaxInheritedRole
				fmt.Println(TranslateRole(myWsAccess.MaxInheritedRole))

				if myWsAccess.NbUsers <= 0 {
					fmt.Println("Accessible to no user")
				} else {
					fmt.Printf("\nAccessible to %d users :\n", myWsAccess.NbUsers)
					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader([]string{"Id", "Nom", "Email", "Inherited access", "Direct access"})
					for _, user := range myWsAccess.Users {
						table.Append([]string{strconv.Itoa(user.Id), user.Name, user.Email, user.ParentAccess, user.Access})
					}
					table.Render()
				}
			}
		case "json":
			{
				jsonAccess, err := json.MarshalIndent(myWsAccess, "", "   ")
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonAccess))
			}
		}
	}
}

// Displays users with access to a document
func DisplayDocAccess(docId string) {
	type UserAccess struct {
		UserId       string `json:"userId"`
		UserEmail    string `json:"userEmail"`
		ParentAccess string `json:"parentAccess"`
		Access       string `json:"access"`
	}
	type DocAcces struct {
		DocId            string       `json:"docId"`
		DocName          string       `json:"docName"`
		WorkspaceName    string       `json:"workspaceName"`
		MaxInheritedRole string       `json:"MaxInheritedRole"`
		UserAccess       []UserAccess `json:"UserAccess"`
	}
	var myDocAccess DocAcces

	// Getting the document
	doc := gristapi.GetDoc(docId)
	if doc.Name == "" {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	} else {
		// Document was found
		// Displaying the access rights
		docAccess := gristapi.GetDocAccess(docId)
		// Sorting users by email (lowercase)
		sort.Slice(docAccess.Users, func(i, j int) bool {
			return strings.ToLower(docAccess.Users[i].Email) < strings.ToLower(docAccess.Users[j].Email)
		})
		var tmpUsers []UserAccess
		for _, user := range docAccess.Users {
			if user.Access != "" {
				userAccess := UserAccess{
					UserId:       strconv.Itoa(user.Id),
					UserEmail:    user.Email,
					ParentAccess: user.ParentAccess,
					Access:       user.Access,
				}

				tmpUsers = append(tmpUsers, userAccess)
			}

			myDocAccess = DocAcces{
				DocId:            doc.Id,
				DocName:          doc.Name,
				WorkspaceName:    doc.Workspace.Name,
				MaxInheritedRole: TranslateRole(docAccess.MaxInheritedRole),
				UserAccess:       tmpUsers,
			}

		}

		switch output {
		case "json":
			{
				jsonData, err := json.MarshalIndent(myDocAccess, "", "   ")
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonData))
			}
		case "table":
			{
				// Displaying the document name
				title := fmt.Sprintf("Workspace \"%s\" (n°%s), document \"%s\"", myDocAccess.WorkspaceName, myDocAccess.DocId, myDocAccess.DocName)
				common.DisplayTitle(title)
				fmt.Println(myDocAccess.MaxInheritedRole)
				fmt.Printf("\nDirect users:\n")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Id", "Email", "Nom", "Inherited access", "Direct access"})
				for _, user := range myDocAccess.UserAccess {
					table.Append([]string{user.UserId, user.UserEmail, user.ParentAccess, user.Access})
				}
				table.Render()
			}
		}
	}
}

// Displaying the rights matrix
func DisplayUserMatrix() {
	type userAccess struct {
		Id            int
		Email         string
		Name          string
		OrgId         int
		OrgName       string
		WorkspaceName string
		WokspaceId    int
		ParentAccess  string
		DirectAccess  string
		Access        string
	}
	lstUserAccess := []userAccess{}

	lstOrg := gristapi.GetOrgs()
	for _, org := range lstOrg {
		for _, ws := range gristapi.GetOrgWorkspaces(org.Id) {
			for _, access := range gristapi.GetWorkspaceAccess(ws.Id).Users {
				tmpUserAccess := userAccess{
					Id:            access.Id,
					Email:         access.Email,
					Name:          access.Name,
					OrgId:         org.Id,
					OrgName:       org.Name,
					WorkspaceName: ws.Name,
					WokspaceId:    ws.Id,
					ParentAccess:  access.ParentAccess,
					DirectAccess:  access.Access,
				}
				if access.Access != "" {
					tmpUserAccess.Access = access.Access
				} else {
					if access.ParentAccess != "" {
						tmpUserAccess.Access = access.Access
					}
				}
				if access.Access != "" {
					lstUserAccess = append(lstUserAccess, tmpUserAccess)
				}
			}
		}
	}

	switch output {
	case "json":
		{
			jsonData, err := json.MarshalIndent(lstUserAccess, "", "   ")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(jsonData))
		}
	case "table":
		{
			accessDf := dataframe.LoadStructs(lstUserAccess)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Id", "Email", "Name", "Org Id", "Org name", "Wokspace id", "Workspace name", "ParentAccess", "DirectAccess", "Access"})
			for email, access := range accessDf.Arrange(dataframe.Sort("Email")).GroupBy("Email").GetGroups() {
				for id, val := range access.Records() {
					if id > 0 {
						line := []string{val[3], email, val[4], val[5], val[6], val[8], val[9], val[7], val[1], val[0]}
						table.Append(line)
					}
				}
			}
			table.Render()
		}
	}
}

// Delete a workspace
func DeleteWorkspace(workspaceId int) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete workspace %d ?", workspaceId)) {
		gristapi.DeleteWorkspace(workspaceId)
	}
}

// Delete a document
func DeleteDoc(docId string) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete document %s ?", docId)) {
		gristapi.DeleteDoc(docId)
	}
}

// Delete a user
func DeleteUser(userId int) {
	if common.Confirm(fmt.Sprintf("Do you really want to delete user %d ?", userId)) {
		gristapi.DeleteUser(userId)
	}
}

// Export a document as a Grist file
func ExportDocGrist(docId string) {
	doc := gristapi.GetDoc(docId)
	if doc.Name != "" {
		gristapi.ExportDocGrist(docId, doc.Workspace.Name+"_"+doc.Name+".grist")
	} else {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	}
}

// Export a document as an Excel file
func ExportDocExcel(docId string) {
	doc := gristapi.GetDoc(docId)
	if doc.Name != "" {
		gristapi.ExportDocExcel(docId, doc.Workspace.Name+"_"+doc.Name+".xlsx")
	} else {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	}
}

// Move a document to a workspace
func MoveDoc(docId string, workspaceId int) {
	doc := gristapi.GetDoc(docId)
	ws := gristapi.GetWorkspace(workspaceId)

	if doc.Name == "" {
		fmt.Printf("❗️ Document %s not found ❗️\n", docId)
	} else {
		if ws.Id == 0 {
			fmt.Printf("❗️ Workspace %d not found ❗️\n", workspaceId)
		} else {
			gristapi.MoveDoc(docId, workspaceId)
		}
	}
}
