package storage

const trace = `
{
    "settings": {
        "index": {
            "number_of_shards": "8",
            "number_of_replicas": "0",
            "routing.allocation.total_shards_per_node": "3"
        }
    },
    "mappings": {
        "_default_": {
            "dynamic_templates": [
                {
                    "strings": {
                        "mapping": {
                            "type": "keyword"
                        },
                        "match_mapping_type": "string",
                        "match": "*"
                    }
                },
                {
                    "value": {
                        "mapping": {
                            "match_mapping_type": "string",
                            "ignore_malformed": true,
                            "type": "keyword"
                        },
                        "match": "value"
                    }
                },
                {
                    "annotations": {
                        "mapping": {
                            "type": "nested"
                        },
                        "match": "annotations"
                    }
                },
                {
                    "binaryAnnotations": {
                        "mapping": {
                            "type": "nested"
                        },
                        "match": "binaryAnnotations"
                    }
                }
            ],
            "_all": {
                "enabled": false
            }
        },
        "span": {
            "dynamic_templates": [
                {
                    "strings": {
                        "mapping": {
                            "type": "keyword"
                        },
                        "match_mapping_type": "string",
                        "match": "*"
                    }
                },
                {
                    "value": {
                        "mapping": {
                            "match_mapping_type": "string",
                            "ignore_malformed": true,
                            "type": "keyword"
                        },
                        "match": "value"
                    }
                },
                {
                    "annotations": {
                        "mapping": {
                            "type": "nested"
                        },
                        "match": "annotations"
                    }
                },
                {
                    "binaryAnnotations": {
                        "mapping": {
                            "type": "nested"
                        },
                        "match": "binaryAnnotations"
                    }
                }
            ],
            "_all": {
                "enabled": false
            },
            "properties": {
                "binaryAnnotations": {
                    "type": "object",
                    "properties": {
                        "endpoint": {
                            "properties": {
                                "ipv4": {
                                    "type": "keyword"
                                },
                                "port": {
                                    "type": "long"
                                },
                                "serviceName": {
                                    "type": "keyword"
                                }
                            }
                        },
                        "value": {
                            "type": "text"
                        },
                        "key": {
                            "type": "keyword"
                        }
                    }
                },
                "duration": {
                    "type": "long"
                },
                "traceId": {
                    "type": "keyword"
                },
                "version": {
                    "type": "keyword"
                },
                "relatedApi": {
                    "type": "keyword"
                },
                "selfApi": {
                    "type": "keyword"
                },
                "name": {
                    "type": "text"
                },
                "annotations": {
                    "type": "object",
                    "properties": {
                        "endpoint": {
                            "properties": {
                                "ipv4": {
                                    "type": "keyword"
                                },
                                "port": {
                                    "type": "long"
                                },
                                "serviceName": {
                                    "type": "keyword"
                                }
                            }
                        },
                        "value": {
                            "type": "keyword"
                        },
                        "timestamp": {
                            "type": "long"
                        }
                    }
                },
                "id": {
                    "type": "keyword"
                },
                "parentId": {
                    "type": "keyword"
                },
                "insertTime": {
                    "type": "date",
                    "format": "yyyy-MM-dd HH:mm:ss"
                },
                "timestamp": {
                    "type": "long"
                }
            }
        }
    }
}
`

var Mappings = map[string]string{
	"trace": trace,
}
