package gristapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"

	"github.com/fatih/color"
)

type Org struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"createdAt"`
}

type Workspace struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Docs      []Doc  `json:"docs"`
}

type Doc struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsPinned bool   `json:"isPinned"`
}

type Table struct {
	Id string `json:"id"`
}

type Tables struct {
	Tables []Table `json:"tables"`
}

type TableColumn struct {
	Id string `json:"id"`
}

type TableColumns struct {
	Columns []TableColumn `json:"columns"`
}

type TableRows struct {
	Id []uint `json:"id"`
}

func get(myRequest string) string {
	// Envoi d'une requête HTTP GET à l'API REST de Grist
	// Retourne le corps de la réponse
	client := &http.Client{}

	url := "https://wpgrist.cus.fr/api/" + myRequest
	bearer := "Bearer " + os.Getenv("TOKEN")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: %s", err)
	}
	req.Header.Add("Authorization", bearer)

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: %s", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("HTTP Error: %s", resp.Status)
	}

	// Read the HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response: %s", err)
	}
	return string(body)
}

func GetOrgs() []Org {
	// Récupère la liste des organisations
	myOrgs := []Org{}
	response := get("orgs")
	json.Unmarshal([]byte(response), &myOrgs)
	return myOrgs
}

func GetOrg(idOrg string) Org {
	// Récupère l'organisation dont l'identifiant est passé en paramètre
	myOrg := Org{}
	response := get("orgs/" + idOrg)
	json.Unmarshal([]byte(response), &myOrg)
	return myOrg
}

func GetOrgWorkspaces(id int) []Workspace {
	lstWorkspaces := []Workspace{}
	response := get("orgs/" + strconv.Itoa(id) + "/workspaces")
	json.Unmarshal([]byte(response), &lstWorkspaces)
	return lstWorkspaces
}

func GetWorkspaceDocs(idWorkspace int) []Doc {
	// Récupère la liste des documents contenus dans un workspace
	lstDocs := []Doc{}
	url := "workspaces/" + strconv.Itoa((idWorkspace)) + "/docs"
	response := get(url)
	json.Unmarshal([]byte(response), &lstDocs)
	return lstDocs
}

func GetDoc(id string) Doc {
	// Récupère un document
	doc := Doc{}
	url := "docs/" + id
	response := get(url)
	json.Unmarshal([]byte(response), &doc)
	return doc
}

func GetDocTables(id string) Tables {
	// Récupère la liste des tables contenues dans un document
	tables := Tables{}
	url := "docs/" + id + "/tables"
	response := get(url)
	json.Unmarshal([]byte(response), &tables)

	return tables
}

func getTableColumns(docId string, columnId string) TableColumns {
	columns := TableColumns{}
	url := "docs/" + docId + "/tables/" + columnId + "/columns"
	response := get(url)
	json.Unmarshal([]byte(response), &columns)

	return columns
}

func GetTableRows(docId string, tableId string) TableRows {
	rows := TableRows{}
	url := "docs/" + docId + "/tables/" + tableId + "/data"
	response := get(url)
	json.Unmarshal([]byte(response), &rows)

	return rows
}

func DisplayDoc(docId string) {
	// Affiche les informations détaillées sur un document
	// - Nom du document
	// - Nombre de tables
	//   Pour chacune d'entre elles :
	//   - Nombre de colonnes
	//   - Nombre de ligne
	//   - Liste des colonnes
	doc := GetDoc(docId)

	type TableDetails struct {
		name       string
		nb_rows    int
		nb_cols    int
		cols_names []string
	}

	title := color.New(color.FgRed).SprintFunc()
	fmt.Printf("\nDocument '%s' (%s)", title(doc.Name), doc.Id)
	if doc.IsPinned {
		fmt.Printf(" - épinglé\n")
	} else {
		fmt.Printf(" - non épinglé\n")
	}

	var tables Tables = GetDocTables(docId)
	fmt.Printf("Contient %d tables\n", len(tables.Tables))
	var wg sync.WaitGroup
	var tables_details []TableDetails
	for _, table := range tables.Tables {
		wg.Add(1)
		go func() {
			defer wg.Done()
			table_desc := ""
			columns := getTableColumns(docId, table.Id)
			rows := GetTableRows(docId, table.Id)

			var cols_names []string
			for _, col := range columns.Columns {
				cols_names = append(cols_names, col.Id)
			}
			slices.Sort(cols_names)
			for _, col := range cols_names {
				table_desc += fmt.Sprintf("%s ", col)
			}
			table_info := TableDetails{
				name:       table.Id,
				nb_rows:    len(rows.Id),
				nb_cols:    len(columns.Columns),
				cols_names: cols_names,
			}
			tables_details = append(tables_details, table_info)
		}()
	}
	wg.Wait()
	var details []string
	for _, table_details := range tables_details {
		ligne := fmt.Sprintf("- \"%s\" : %d lignes, %d colonnes\n", table_details.name, table_details.nb_rows, table_details.nb_cols)
		for _, col_name := range table_details.cols_names {
			ligne = ligne + fmt.Sprintf("  - %s\n", col_name)
		}
		details = append(details, ligne)
	}
	sort.Strings(details)
	for _, ligne := range details {
		fmt.Printf("%s", ligne)
	}
}

func DisplayOrgs() {
	for _, org := range GetOrgs() {
		fmt.Printf("- Org n°%d (%s)\n", org.Id, org.Name)
	}
}

func DisplayOrg(orgId string) {
	// Affiche les détails sur une organisation
	org := GetOrg(orgId)
	fmt.Printf("Organisation n°%d : %s\n", org.Id, org.Name)
	worskspaces := GetOrgWorkspaces(org.Id)
	fmt.Printf("  %d workspaces : \n", len(worskspaces))
	for _, ws := range worskspaces {
		if len(ws.Docs) > 0 {
			var lst_docs []string
			for _, doc := range ws.Docs {
				lst_docs = append(lst_docs, fmt.Sprintf("%s (%s)", doc.Name, doc.Id))
			}
			slices.Sort(lst_docs)
			fmt.Printf("  - \"%s\" (n°%d), %d documents:\n", ws.Name, ws.Id, len(ws.Docs))
			for _, doc := range lst_docs {
				fmt.Printf("    - %s\n", doc)
			}
		}
	}
}