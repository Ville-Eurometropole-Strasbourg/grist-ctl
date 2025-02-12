// SPDX-FileCopyrightText: 2024 Ville Eurométropole Strasbourg
//
// SPDX-License-Identifier: MIT

// Main program
package main

import (
	"fmt"
	"os"
	"strconv"

	"gristctl/gristapi"
	"gristctl/gristtools"
)

var version = "Undefined"

func main() {
	args := os.Args[1:]
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
								gristapi.ExportDocGrist(docId)
							case "excel":
								gristapi.ExportDocExcel(docId)
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
	}

}
