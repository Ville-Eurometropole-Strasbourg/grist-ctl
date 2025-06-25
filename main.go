// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Main program
package main

import (
	"flag"
	"fmt"
	"gristctl/gristapi"
	"gristctl/gristtools"
	"strconv"
)

var version = "Undefined"

func main() {
	// Define the options
	optionOutput := flag.String("o", "table", "Output format")

	flag.Parse()

	switch *optionOutput {
	case "json":
		gristtools.SetOutput("json")
	default:
		gristtools.SetOutput("table")
	}

	args := flag.Args()

	if len(args) < 1 {
		gristtools.Help()
	}

	switch arg1 := args[0]; arg1 {
	case "config":
		gristtools.Config()
	case "version":
		gristtools.Version(version)
	case "get":
		{
			if len(args) > 1 {
				switch arg2 := args[1]; arg2 {
				case "org":
					{
						switch nb := len(args); nb {
						case 2:
							gristtools.DisplayOrgs()
						case 3:
							orgId := args[2]
							gristtools.DisplayOrg(orgId)
						case 4:
							switch args[3] {
							case "access":
								orgId := args[2]
								gristtools.DisplayOrgAccess(orgId)
							default:
								gristtools.Help()
							}
						default:
							gristtools.Help()
							flag.PrintDefaults()
						}
					}
				case "doc":
					{
						switch len(args) {
						case 3:
							docId := args[2]
							gristtools.DisplayDoc(docId)
						case 4:
							docId := args[2]
							switch args[3] {
							case "access":
								gristtools.DisplayDocAccess(docId)
							case "grist":
								gristtools.ExportDocGrist(docId)
							case "excel":
								gristtools.ExportDocExcel(docId)
							default:
								fmt.Println("You have to choose between 'access', 'grist', or 'excel'")
							}
						case 5:
							docId := args[2]
							switch args[3] {
							case "table":
								tableName := args[4]
								gristapi.GetTableContent(docId, tableName)
							default:
								gristtools.Help()
							}

						default:
							gristtools.Help()
						}
					}
				case "workspace":
					{
						switch len(args) {
						case 3:
							worskspaceId, err := strconv.Atoi(args[2])
							if err == nil {
								gristtools.DisplayWorkspace(worskspaceId)
							}
						case 4:
							if args[3] == "access" {
								worskspaceId, err := strconv.Atoi(args[2])
								if err == nil {
									gristtools.DisplayWorkspaceAccess(worskspaceId)
								}
							}
						default:
							gristtools.Help()
						}
					}
				case "users":
					gristtools.DisplayUserMatrix()
				default:
					gristtools.Help()
				}
			}
		}
	case "move":
		{
			if len(args) > 2 {
				switch args[1] {
				case "doc":
					docId := args[2]
					switch args[3] {
					case "workspace":
						workspaceId := 0
						id, err := strconv.Atoi(args[4])
						if err == nil {
							workspaceId = id
						} else {
							gristtools.Help()
						}
						gristapi.MoveDoc(docId, workspaceId)
					}
				case "docs":
					fromWorkspaceId := 0
					toWorkspaceId := 0
					fromId, err := strconv.Atoi(args[3])
					if err == nil {
						fromWorkspaceId = fromId
					} else {
						gristtools.Help()
					}
					toId, err := strconv.Atoi(args[5])
					if err == nil {
						toWorkspaceId = toId
					} else {
						gristtools.Help()
					}
					gristapi.MoveAllDocs(fromWorkspaceId, toWorkspaceId)
				default:
					gristtools.Help()
				}
			}
		}
	case "purge":
		{
			if len(args) > 2 {
				switch args[1] {
				case "doc":
					docId := args[2]
					nbHisto := 3
					if len(args) == 4 {
						nb, err := strconv.Atoi(args[3])
						if err == nil {
							nbHisto = nb
						} else {
							gristtools.Help()
						}
					}
					gristapi.PurgeDoc(docId, nbHisto)
				default:
					gristtools.Help()
				}
			}
		}
	case "delete":
		{
			if len(args) > 2 {
				switch arg2 := args[1]; arg2 {
				case "workspace":
					if len(args) == 3 {
						idWorkspace, err := strconv.Atoi(args[2])
						if err == nil {
							gristtools.DeleteWorkspace(idWorkspace)
						}
					} else {
						gristtools.Help()
					}
				case "user":
					if len(args) == 3 {
						idUser, err := strconv.Atoi(args[2])
						if err == nil {
							gristtools.DeleteUser(idUser)
						}
					} else {
						gristtools.Help()
					}
				case "doc":
					if len(args) == 3 {
						docId := args[2]
						gristtools.DeleteDoc(docId)
					}
				default:
					gristtools.Help()
				}
			}
		}
	case "import":
		if len(args) > 1 {
			switch args[1] {
			case "users":
				gristtools.ImportUsers()
			default:
				gristtools.Help()
			}
		}
	default:
		gristtools.Help()
		flag.PrintDefaults()
	}

}
