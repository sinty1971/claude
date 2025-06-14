// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/file-entries": {
            "get": {
                "description": "Retrieve a list of folders from the specified path",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "file-entries"
                ],
                "summary": "Get folders",
                "parameters": [
                    {
                        "type": "string",
                        "default": "~/penguin",
                        "description": "Path to the directory to list",
                        "name": "path",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response",
                        "schema": {
                            "$ref": "#/definitions/models.FolderListResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/kouji-entries": {
            "get": {
                "description": "指定されたパスから工事プロジェクトフォルダーの一覧を取得します。\n各工事プロジェクトには会社名、現場名、工事開始日などの詳細情報が含まれます。",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "工事管理"
                ],
                "summary": "工事プロジェクト一覧の取得",
                "parameters": [
                    {
                        "type": "string",
                        "default": "~/penguin/豊田築炉/2-工事",
                        "description": "工事フォルダーのパス",
                        "name": "path",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "工事プロジェクト一覧",
                        "schema": {
                            "$ref": "#/definitions/models.KoujiEntriesResponse"
                        }
                    },
                    "500": {
                        "description": "サーバーエラー",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/kouji-entries/cleanup": {
            "post": {
                "description": "Remove kouji entries with invalid time data (like 0001-01-01T09:26:51+09:18) from YAML",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "kouji-entries"
                ],
                "summary": "Cleanup invalid time data",
                "parameters": [
                    {
                        "type": "string",
                        "default": "~/penguin/豊田築炉/2-工事/.inside.yaml",
                        "description": "Path to the YAML file",
                        "name": "yaml_path",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message with cleanup details",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/kouji-entries/save": {
            "post": {
                "description": "Save kouji entries information to a YAML file",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "kouji-entries"
                ],
                "summary": "Save kouji entries to YAML",
                "parameters": [
                    {
                        "type": "string",
                        "default": "~/penguin/豊田築炉/2-工事",
                        "description": "Path to the directory to scan",
                        "name": "path",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "~/penguin/豊田築炉/2-工事/.inside.yaml",
                        "description": "Output YAML file path",
                        "name": "output_path",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/kouji-entries/{project_id}/dates": {
            "put": {
                "description": "Update start and end dates for a specific kouji project",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "kouji-entries"
                ],
                "summary": "Update kouji project dates",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Project ID",
                        "name": "project_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated dates",
                        "name": "dates",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UpdateKoujiEntryDatesRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/time/formats": {
            "get": {
                "description": "Get list of all supported date/time formats",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "time"
                ],
                "summary": "Get supported time formats",
                "responses": {
                    "200": {
                        "description": "Successful response",
                        "schema": {
                            "$ref": "#/definitions/models.SupportedFormatsResponse"
                        }
                    }
                }
            }
        },
        "/time/parse": {
            "post": {
                "description": "Parse various date/time string formats",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "time"
                ],
                "summary": "Parse time string",
                "parameters": [
                    {
                        "description": "Time string to parse",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.TimeParseRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response",
                        "schema": {
                            "$ref": "#/definitions/models.TimeParseResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.FileEntry": {
            "description": "File or directory information",
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 123456
                },
                "is_directory": {
                    "description": "Whether this item is a directory",
                    "type": "boolean",
                    "example": true
                },
                "modified_time": {
                    "description": "Last modification time",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.Timestamp"
                        }
                    ]
                },
                "name": {
                    "description": "Name of the file or folder",
                    "type": "string",
                    "example": "documents"
                },
                "path": {
                    "description": "Full path to the file or folder",
                    "type": "string",
                    "example": "/home/user/documents"
                },
                "size": {
                    "description": "Size of the file in bytes",
                    "type": "integer",
                    "example": 4096
                }
            }
        },
        "models.FolderListResponse": {
            "description": "Response containing list of folders",
            "type": "object",
            "properties": {
                "count": {
                    "description": "Total number of folders returned",
                    "type": "integer",
                    "example": 10
                },
                "folders": {
                    "description": "List of folders",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.FileEntry"
                    }
                }
            }
        },
        "models.KoujiEntriesResponse": {
            "description": "Response containing list of construction kouji folders",
            "type": "object",
            "properties": {
                "count": {
                    "type": "integer",
                    "example": 10
                },
                "kouji_entries": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.KoujiEntry"
                    }
                },
                "total_size": {
                    "type": "integer",
                    "example": 1073741824
                }
            }
        },
        "models.KoujiEntry": {
            "description": "Construction kouji folder information with extended attributes",
            "type": "object",
            "properties": {
                "company_name": {
                    "type": "string",
                    "example": "豊田築炉"
                },
                "description": {
                    "type": "string",
                    "example": "工事関連の資料とドキュメント"
                },
                "end_date": {
                    "$ref": "#/definitions/models.Timestamp"
                },
                "file_count": {
                    "type": "integer",
                    "example": 42
                },
                "id": {
                    "type": "integer",
                    "example": 123456
                },
                "is_directory": {
                    "description": "Whether this item is a directory",
                    "type": "boolean",
                    "example": true
                },
                "location_name": {
                    "type": "string",
                    "example": "名和工場"
                },
                "modified_time": {
                    "description": "Last modification time",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.Timestamp"
                        }
                    ]
                },
                "name": {
                    "description": "Name of the file or folder",
                    "type": "string",
                    "example": "documents"
                },
                "path": {
                    "description": "Full path to the file or folder",
                    "type": "string",
                    "example": "/home/user/documents"
                },
                "size": {
                    "description": "Size of the file in bytes",
                    "type": "integer",
                    "example": 4096
                },
                "start_date": {
                    "$ref": "#/definitions/models.Timestamp"
                },
                "status": {
                    "type": "string",
                    "example": "進行中"
                },
                "subdir_count": {
                    "type": "integer",
                    "example": 5
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "['工事'",
                        " '豊田築炉'",
                        " '名和工場']"
                    ]
                }
            }
        },
        "models.SupportedFormatsResponse": {
            "description": "List of all supported date/time formats",
            "type": "object",
            "properties": {
                "formats": {
                    "description": "List of supported formats",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.TimeFormat"
                    }
                }
            }
        },
        "models.TimeFormat": {
            "description": "Supported time format information",
            "type": "object",
            "properties": {
                "example": {
                    "description": "Example value",
                    "type": "string",
                    "example": "2024-01-15T10:30:00Z"
                },
                "name": {
                    "description": "Format name",
                    "type": "string",
                    "example": "RFC3339"
                },
                "pattern": {
                    "description": "Format pattern",
                    "type": "string",
                    "example": "2006-01-02T15:04:05Z07:00"
                }
            }
        },
        "models.TimeParseRequest": {
            "description": "Request for parsing various date/time formats",
            "type": "object",
            "properties": {
                "time_string": {
                    "description": "Time string to parse",
                    "type": "string",
                    "example": "2024-01-15T10:30:00"
                }
            }
        },
        "models.TimeParseResponse": {
            "description": "Response containing parsed time in various formats",
            "type": "object",
            "properties": {
                "original": {
                    "description": "Original input string",
                    "type": "string",
                    "example": "2024-01-15T10:30:00"
                },
                "readable": {
                    "description": "Human readable format",
                    "type": "string",
                    "example": "January 15, 2024 10:30 AM"
                },
                "rfc3339": {
                    "description": "Parsed time in RFC3339 format",
                    "type": "string",
                    "example": "2024-01-15T10:30:00Z"
                },
                "timezone": {
                    "description": "Time zone used",
                    "type": "string",
                    "example": "Local"
                },
                "unix": {
                    "description": "Unix timestamp",
                    "type": "integer",
                    "example": 1705318200
                }
            }
        },
        "models.Timestamp": {
            "description": "Timestamp in RFC3339 format",
            "type": "object",
            "properties": {
                "time.Time": {
                    "type": "string"
                }
            }
        },
        "models.UpdateKoujiEntryDatesRequest": {
            "description": "Request body for updating kouji start and end dates",
            "type": "object",
            "properties": {
                "end_date": {
                    "type": "string",
                    "example": "2024-12-31T00:00:00Z"
                },
                "start_date": {
                    "type": "string",
                    "example": "2024-01-01T00:00:00Z"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "Penguin FileSystem Management API",
	Description:      "API for managing and browsing file entries",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
