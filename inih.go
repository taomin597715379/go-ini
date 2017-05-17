package ini

import (
	"io"
)

const DEFAULT_SECTION = "default"

type ini_break_t int

// Line break types.
const (
	// Let the parser choose the break type.
	ini_ANY_BREAK ini_break_t = iota

	ini_CR_BREAK   // Use CR for line breaks (Mac style).
	ini_LN_BREAK   // Use LN for line breaks (Unix style).
	ini_CRLN_BREAK // Use CR LN for line breaks (DOS style).
)

type ini_error_type_t int

// Many bad things could happen with the parser and emitter.
const (
	// No error is produced.
	ini_NO_ERROR ini_error_type_t = iota

	ini_MEMORY_ERROR   // Cannot allocate or reallocate a block of memory.
	ini_READER_ERROR   // Cannot read or decode the input stream.
	ini_SCANNER_ERROR  // Cannot scan the input stream.
	ini_PARSER_ERROR   // Cannot parse the input stream.
	ini_COMPOSER_ERROR // Cannot compose a YAML document.
	ini_WRITER_ERROR   // Cannot write to the output stream.
	ini_EMITTER_ERROR  // Cannot emit a YAML stream.
)

// The pointer position.
type ini_mark_t struct {
	index int // The position index.
	line  int // The position line.
    column int // The position column.
}

// Node Styles

type ini_style_t int8

type ini_scalar_style_t ini_style_t

// Scalar styles.
const (
	// Let the emitter choose the style.
	ini_ANY_SCALAR_STYLE ini_scalar_style_t = iota

    ini_PLAIN_SCALAR_STYLE       // The literal scalar style.
    ini_LITERAL_SCALAR_STYLE       // The literal scalar style.
	ini_SINGLE_QUOTED_SCALAR_STYLE // The single-quoted scalar style.
	ini_DOUBLE_QUOTED_SCALAR_STYLE // The double-quoted scalar style.
)

// Tokens

type ini_token_type_t int

// Token types.
const (
	// An empty token.
	ini_NO_TOKEN ini_token_type_t = iota

	ini_STREAM_START_TOKEN   // A DOCUMENT-START token.
	ini_STREAM_END_TOKEN     // A DOCUMENT-START token.
	ini_DOCUMENT_START_TOKEN // A DOCUMENT-START token.
	ini_DOCUMENT_END_TOKEN   // A DOCUMENT-END token.

	ini_SECTION_START_TOKEN   // A SECTION-START token.
	ini_SECTION_END_TOKEN     // A SECTION-END token.
	ini_SECTION_INHERIT_TOKEN // A SECTION-INHERIT token.

	ini_KEY_TOKEN   // An NODE KEY token.
	ini_VALUE_TOKEN // An NODE VALUE token.

	ini_COMMENT_START_TOKEN // A COMMENT-START token.
	ini_COMMENT_END_TOKEN   // A COMMENT-END token.
)

func (tt ini_token_type_t) String() string {
	switch tt {
	case ini_NO_TOKEN:
		return "ini_NO_TOKEN"
	case ini_STREAM_START_TOKEN:
		return "ini_STREAM_START_TOKEN"
	case ini_STREAM_END_TOKEN:
		return "ini_STREAM_END_TOKEN"
	case ini_DOCUMENT_START_TOKEN:
		return "ini_DOCUMENT_START_TOKEN"
	case ini_DOCUMENT_END_TOKEN:
		return "ini_DOCUMENT_END_TOKEN"
	case ini_SECTION_START_TOKEN:
		return "ini_SECTION_START_TOKEN"
	case ini_SECTION_END_TOKEN:
		return "ini_SECTION_END_TOKEN"
	case ini_SECTION_INHERIT_TOKEN:
		return "ini_SECTION_INHERIT_TOKEN"
	case ini_KEY_TOKEN:
		return "ini_KEY_TOKEN"
	case ini_VALUE_TOKEN:
		return "ini_VALUE_TOKEN"
	case ini_COMMENT_START_TOKEN:
		return "ini_COMMENT_START_TOKEN"
	case ini_COMMENT_END_TOKEN:
		return "ini_COMMENT_END_TOKEN"
	}
	return "<unknown token>"
}

// The token structure.
type ini_token_t struct {
	// The token type.
	typ ini_token_type_t

	// The start/end of the token.
	start_mark, end_mark ini_mark_t

	// The scalar value
	// (for ini_SCALAR_TOKEN).
	value []byte

	// The scalar style (for ini_VALUE_TOKEN).
	style ini_scalar_style_t
}

// Events

type ini_event_type_t int8

// Event types.
const (
	// An empty event.
	ini_NO_EVENT ini_event_type_t = iota

	ini_STREAM_START_EVENT    // A STREAM-START event.
	ini_STREAM_END_EVENT      // A STREAM-END event.
	ini_DOCUMENT_START_EVENT  // A DOCUMENT-START event.
	ini_DOCUMENT_END_EVENT    // A DOCUMENT-END event.
	ini_SECTION_START_EVENT   // A SECTION-START event.
	ini_SECTION_INHERIT_EVENT // A INHERIT SECTION event.
	ini_SECTION_END_EVENT     // A SECTION-END event.

	ini_SCALAR_EVENT  // An SCALAR event.
	ini_KEY_EVENT     // An KEY event.
	ini_VALUE_EVENT   // An VALUE event.
	ini_COMMENT_EVENT // A COMMENT event.
)

// The event structure.
type ini_event_t struct {

	// The event type.
	typ ini_event_type_t

	// The start and end of the event.
	start_mark, end_mark ini_mark_t

	// The node value (for ini_NODE_EVENT).
	value []byte

	// for ini_NODE_EVENT.
	implicit bool

	// The style (for ini_NODE_EVENT).
	style ini_style_t
}

func (e *ini_event_t) scalar_style() ini_scalar_style_t { return ini_scalar_style_t(e.style) }

type ini_node_type_t int

// Node types.
const (
	// An empty node.
	ini_NO_NODE ini_node_type_t = iota

	ini_ITEM_KEY_NODE   // A item key node.
	ini_ITEM_VALUE_NODE // A item value node.
	ini_COMMENT_NODE    // A comment node.
)

// The node structure.
type ini_node_t struct {
	typ ini_node_type_t // The node type.

	// The node data.

	// The scalar parameters (for ini_SCALAR_NODE).
	scalar struct {
		value  []byte             // The scalar value.
		length int                // The length of the scalar value.
		style  ini_scalar_style_t // The scalar style.
	}

	start_mark ini_mark_t // The beginning of the node.
	end_mark   ini_mark_t // The end of the node.

}

// The document structure.
type ini_document_t struct {

	// The document nodes.
	nodes []ini_node_t

	// The start/end of the document.
	start_mark, end_mark ini_mark_t
}

// The prototype of a read handler.
//
// The read handler is called when the parser needs to read more bytes from the
// source. The handler should write not more than size bytes to the buffer.
// The number of written bytes should be set to the size_read variable.
//
// [in,out]   data        A pointer to an application data specified by
//                        ini_parser_set_input().
// [out]      buffer      The buffer to write the data from the source.
// [in]       size        The size of the buffer.
// [out]      size_read   The actual number of bytes read from the source.
//
// On success, the handler should return 1.  If the handler failed,
// the returned value should be 0. On EOF, the handler should set the
// size_read to 0 and return 1.
type ini_read_handler_t func(parser *ini_parser_t, buffer []byte) (n int, err error)

// The states of the parser.
type ini_parser_state_t int

const (
	ini_PARSE_STREAM_START_STATE ini_parser_state_t = iota // Expect START.

	ini_PARSE_DOCUMENT_START_STATE      // Expect DOCUMENT-START.
	ini_PARSE_DOCUMENT_END_STATE        // Expect DOCUMENT-END.
    ini_PARSE_DOCUMENT_CONTENT_STATE    // Expect DOCUMENT-CONTENT.
	ini_PARSE_SECTION_FIRST_ENTRY_STATE // Expect SECTION-ENTRY.
	ini_PARSE_SECTION_ENTRY_STATE       // Expect SECTION-ENTRY.
	ini_PARSE_COMMENT_START_STATE       // Expect COMMENT-START.
	ini_PARSE_COMMENT_CONTENT_STATE     // Expect the content of a comment.
	ini_PARSE_COMMENT_END_STATE         // Expect COMMENT-END.
	ini_PARSE_KEY_STATE                 // Expect a node key.
	ini_PARSE_VALUE_STATE               // Expect a node value.
	ini_PARSE_STREAM_END_STATE          // Expect END.
)

func (ps ini_parser_state_t) String() string {
	switch ps {
	case ini_PARSE_STREAM_START_STATE:
		return "ini_PARSE_STREAM_START_STATE"
	case ini_PARSE_STREAM_END_STATE:
		return "ini_PARSE_DOCUMENT_END_STATE"
	case ini_PARSE_DOCUMENT_START_STATE:
		return "ini_PARSE_DOCUMENT_START_STATE"
    case ini_PARSE_DOCUMENT_CONTENT_STATE:
        return "ini_PARSE_DOCUMENT_CONTENT_STATE"
	case ini_PARSE_DOCUMENT_END_STATE:
		return "ini_PARSE_DOCUMENT_END_STATE"
	case ini_PARSE_SECTION_ENTRY_STATE:
		return "ini_PARSE_SECTION_ENTRY_STATE"
	case ini_PARSE_COMMENT_START_STATE:
		return "ini_PARSE_COMMENT_START_STATE"
	case ini_PARSE_COMMENT_CONTENT_STATE:
		return "ini_PARSE_COMMENT_CONTENT_STATE"
	case ini_PARSE_COMMENT_END_STATE:
		return "ini_PARSE_COMMENT_END_STATE"
	case ini_PARSE_KEY_STATE:
		return "ini_PARSE_KEY_STATE"
	case ini_PARSE_VALUE_STATE:
		return "ini_PARSE_VALUE_STATE"
	}
	return "<unknown parser state>"
}

// The parser structure.
//
// All members are internal. Manage the structure using the
// ini_parser_ family of functions.
type ini_parser_t struct {

	// Error handling
	error ini_error_type_t // Error type.

	problem string // Error description.

	// The byte about which the problem occured.
	problem_offset int
	problem_value  int
	problem_mark   ini_mark_t

	// The error context.
	context      string
	context_mark ini_mark_t

	// Reader stuff
	read_handler ini_read_handler_t // Read handler.

	input_file io.Reader // File input data.
	input      []byte    // String input data.
	input_pos  int

	eof bool // EOF flag

	buffer     []byte // The working buffer.
	buffer_pos int    // The current position of the buffer.

	unread int // The number of unread characters in the buffer.

	raw_buffer     []byte // The raw buffer.
	raw_buffer_pos int    // The current position of the buffer.

	offset int        // The offset of the current position (in bytes).
	mark   ini_mark_t // The mark of the current position.

	level int // The current flow level.

	// Scanner stuff
	stream_start_produced bool // Have we started to scan the input stream?
	stream_end_produced   bool // Have we reached the end of the input stream?

	tokens          []ini_token_t // The tokens queue.
	tokens_head     int           // The head of the tokens queue.
	tokens_parsed   int           // The number of tokens fetched from the queue.
	token_available bool          // Does the tokens queue contain a token ready for dequeueing.

	// Parser stuff
	state  ini_parser_state_t   // The current parser state.
	states []ini_parser_state_t // The parser states stack.
	marks  []ini_mark_t         // The stack of marks.
}

// Emitter Definitions

// The prototype of a write handler.
//
// The write handler is called when the emitter needs to flush the accumulated
// characters to the output.  The handler should write @a size bytes of the
// @a buffer to the output.
//
// @param[in,out]   data        A pointer to an application data specified by
//                              ini_emitter_set_output().
// @param[in]       buffer      The buffer with bytes to be written.
// @param[in]       size        The size of the buffer.
//
// @returns On success, the handler should return @c 1.  If the handler failed,
// the returned value should be @c 0.
//
type ini_write_handler_t func(emitter *ini_emitter_t, buffer []byte) error

type ini_emitter_state_t int

// The emitter states.
const (
	// Expect STREAM-START.
	ini_EMIT_STREAM_START_STATE ini_emitter_state_t = iota

	ini_EMIT_DOCUMENT_START_STATE         // Expect DOCUMENT-START.
	ini_EMIT_DOCUMENT_END_STATE           // Expect DOCUMENT-END.
	ini_EMIT_FIRST_SECTION_START_STATE    // Expect the first section
	ini_EMIT_SECTION_START_STATE          // Expect the start of section.
	ini_EMIT_SECTION_FIRST_NODE_KEY_STATE // Expect the start of section.
	ini_EMIT_SECTION_NODE_KEY_STATE       // Expect the start of section.
	ini_EMIT_SECTION_NODE_VALUE_STATE     // Expect the node.
	ini_EMIT_SECTION_END_STATE            // Expect the end of section.
	ini_EMIT_COMMENT_START_STATE          // Expect the start of section.
	ini_EMIT_COMMENT_VALUE_STATE          // Expect the content of section.
	ini_EMIT_COMMENT_END_STATE            // Expect the end of section.
	ini_EMIT_STREAM_END_STATE             // Expect the end of section.

)

// The emitter structure.
//
// All members are internal.  Manage the structure using the @c ini_emitter_
// family of functions.
type ini_emitter_t struct {

	// Error handling

	error   ini_error_type_t // Error type.
	problem string           // Error description.

	// Writer stuff

	write_handler ini_write_handler_t // Write handler.

	output_buffer *[]byte   // String output data.
	output_file   io.Writer // File output data.

	buffer     []byte // The working buffer.
	buffer_pos int    // The current position of the buffer.

	raw_buffer     []byte // The raw buffer.
	raw_buffer_pos int    // The current position of the buffer.

	// Emitter stuff

	unicode    bool        // Allow unescaped non-ASCII characters?
	line_break ini_break_t // The preferred line break.

	state  ini_emitter_state_t   // The current emitter state.
	states []ini_emitter_state_t // The stack of states.

	events      []ini_event_t // The event queue.
	events_head int           // The head of the event queue.

	level int // The current flow level.

	root_context    bool // Is it the document root context?
	mapping_context bool // Is it a mapping context?

	line       int  // The current line.
	column     int  // The current column.
	whitespace bool // If the last character was a whitespace?
	open_ended bool // If an explicit document end is required?

	// Scalar analysis.
	scalar_data struct {
		value                 []byte             // The scalar value.
		multiline             bool               // Does the scalar contain line breaks?
		single_quoted_allowed bool               // Can the scalar be expressed in the single quoted style?
		style                 ini_scalar_style_t // The output style.
	}

	// Dumper stuff

	opened bool // If the document was already opened?
	closed bool // If the document was already closed?
}