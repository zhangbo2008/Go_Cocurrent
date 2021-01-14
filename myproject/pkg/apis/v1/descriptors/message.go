package descriptors

import (
	"awesomeProject8/myproject/pkg/handlers"

	def "github.com/caicloud/nirvana/definition"
)

func init() {
	register([]def.Descriptor{{
		Path:        "/messages",
		Definitions: []def.Definition{listMessages},
	}, {
		Path:        "/messages/{message}",
		Definitions: []def.Definition{getMessage},
	},
	}...)
}

var listMessages = def.Definition{
	Method:      def.List,
	Summary:     "List Messages",
	Description: "Query a specified number of messages and returns an array",
	Function:    handlers.ListMessages,
	Parameters: []def.Parameter{
		{
			Source:      def.Query,
			Name:        "count",
			Default:     2,
			Description: "Number of messages",
		},
	},
	Results: def.DataErrorResults("A list of messages"),
}

var getMessage = def.Definition{
	Method:      def.Get,
	Summary:     "Get Message",
	Description: "Get a message by id",
	Function:    handlers.GetMessage,
	Parameters: []def.Parameter{
		def.PathParameterFor("message", "Message id"),
	},
	Results: def.DataErrorResults("A message"),
}
