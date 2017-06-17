package ini

// The parser implements the following grammar:
//
// document		::= DOCUMENT-START section* DOCUMENT-END
// section      ::= SECTION-START (node | comment)* SECTION-END
// node         ::= KEY VALUE SCALAR
// comment      ::= COMMENT SCALAR

// Peek the next token in the token queue.
func peek_token(parser *ini_parser_t) *ini_token_t {
	if parser.token_available || ini_parser_fetch_more_tokens(parser) {
		return &parser.tokens[parser.tokens_head]
	}
	return nil
}

// Remove the next token from the queue (must be called after peek_token).
func skip_token(parser *ini_parser_t) {
	parser.token_available = false
	parser.tokens_parsed++
	parser.document_end_produced = parser.tokens[parser.tokens_head].typ == ini_DOCUMENT_END_TOKEN
	parser.tokens_head++
}

// Get the next event.
func ini_parser_parse(parser *ini_parser_t, event *ini_event_t) bool {
	// Erase the event object.
	*event = ini_event_t{}

	// No events after the end of the stream or error.
	if parser.document_end_produced || parser.error != ini_NO_ERROR || parser.state == ini_PARSE_DOCUMENT_END_STATE {
		return true
	}

	// Generate the next event.
	return ini_parser_state_machine(parser, event)
}

// Set parser error.
func ini_parser_set_parser_error(parser *ini_parser_t, problem string, problem_mark ini_mark_t) bool {
	parser.error = ini_PARSER_ERROR
	parser.problem = problem
	parser.problem_mark = problem_mark
	return false
}

func ini_parser_set_parser_error_context(parser *ini_parser_t, context string, context_mark ini_mark_t, problem string, problem_mark ini_mark_t) bool {
	parser.error = ini_PARSER_ERROR
	parser.context = context
	parser.context_mark = context_mark
	parser.problem = problem
	parser.problem_mark = problem_mark
	return false
}

// State dispatcher.
func ini_parser_state_machine(parser *ini_parser_t, event *ini_event_t) bool {
	//trace("ini_parser_state_machine", "state:", parser.state.String())
	switch parser.state {
	case ini_PARSE_DOCUMENT_START_STATE:
		return ini_parser_parse_document_start(parser, event)
	case ini_PARSE_SECTION_FIRST_ENTRY_STATE:
		return ini_parser_parse_section_entry(parser, event, true)
	case ini_PARSE_SECTION_ENTRY_STATE:
		return ini_parser_parse_section_entry(parser, event, false)
	case ini_PARSE_SECTION_FIRST_KEY_STATE:
		return ini_parser_parse_section_key(parser, event, true)
	case ini_PARSE_SECTION_KEY_STATE:
		return ini_parser_parse_section_key(parser, event, false)
	case ini_PARSE_SECTION_VALUE_STATE:
		return ini_parser_parse_section_value(parser, event)
	default:
		panic("invalid parser state")
	}
	return false
}

func ini_parser_parse_document_start(parser *ini_parser_t, event *ini_event_t) bool {
	token := peek_token(parser)
	if token == nil {
		return false
	}
	skip_token(parser)
	if token.typ != ini_DOCUMENT_START_TOKEN {
		return ini_parser_set_parser_error(parser, "did not find expected <document-start>", token.start_mark)
	}
	parser.states = append(parser.states, ini_PARSE_DOCUMENT_END_STATE)
	parser.state = ini_PARSE_SECTION_FIRST_ENTRY_STATE
	*event = ini_event_t{
		typ:        ini_DOCUMENT_START_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.end_mark,
	}
	return true
}

// Parse the section:
func ini_parser_parse_section_entry(parser *ini_parser_t, event *ini_event_t, first bool) bool {
	//defer trace("ini_parser_parse_section_entry")

	token := peek_token(parser)
	if token == nil {
		return false
	}
	if token.typ == ini_DOCUMENT_END_TOKEN {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		*event = ini_event_t{
			typ:        ini_DOCUMENT_END_EVENT,
			start_mark: token.start_mark,
			end_mark:   token.end_mark,
		}
		return true
	} else {
		if first && token.typ == ini_SECTION_KEY_TOKEN {
			parser.state = ini_PARSE_SECTION_FIRST_KEY_STATE
			*event = ini_event_t{
				typ:        ini_SECTION_ENTRY_EVENT,
				start_mark: token.start_mark,
				end_mark:   token.start_mark,
				value:      []byte("default"),
				tag:        []byte(ini_SECTION_TAG),
			}
			return true
		} else if token.typ == ini_SECTION_START_TOKEN {
            start_mark := token.start_mark
			skip_token(parser)
			token := peek_token(parser)
			if token == nil {
				return false
			}
			skip_token(parser)
			if token.typ == ini_SCALAR_TOKEN {
                section_name := token.value[0:]
                inherit := false
                token := peek_token(parser)
                if token == nil {
                    return false
                }
                end_mark := token.end_mark
                // SECTION-INHERIT Token (:)
                if token.typ == ini_SECTION_INHERIT_TOKEN {
                    inherit = true
                    skip_token(parser)
                    for _, v := range token.value {
                        section_name = append(section_name, v)
                    }
                    token = peek_token(parser)
                    if token == nil {
                        return false
                    }
                    skip_token(parser)
                    for _, v := range token.value {
                        section_name = append(section_name, v)
                    }
                    end_mark =  token.end_mark
                    if token.typ == ini_SCALAR_TOKEN {
                        parser.state = ini_PARSE_SECTION_FIRST_KEY_STATE
                    } else {
                        return ini_parser_set_parser_error(parser, "did not find expected <scalar>", token.start_mark)
                    }
                    token = peek_token(parser)
                    if token == nil {
                        return false
                    }
                }
                if token.typ == ini_SECTION_END_TOKEN {
                    skip_token(parser)
                    parser.state = ini_PARSE_SECTION_FIRST_KEY_STATE
                    if inherit {
                        *event = ini_event_t{
                            typ:        ini_SECTION_ENTRY_EVENT,
                            start_mark: start_mark,
                            end_mark:   end_mark,
                            value:      section_name,
                            tag:        []byte(ini_SECTION_INHERIT_TAG),
                            style:      ini_style_t(token.style),
                        }
                    } else {
                        *event = ini_event_t{
                            typ:        ini_SECTION_ENTRY_EVENT,
                            start_mark: start_mark,
                            end_mark:   end_mark,
                            value:      section_name,
                            tag:        []byte(ini_SECTION_TAG),
                            style:      ini_style_t(token.style),
                        }
                    }
                    return true
                } else {
                    return ini_parser_set_parser_error(parser, "did not find expected <section-end>", token.start_mark)
                }
			} else {
                return ini_parser_set_parser_error(parser, "did not find expected <scalar>", token.start_mark)
            }
		} else if token.typ == ini_SECTION_END_TOKEN {
            skip_token(parser)
            token = peek_token(parser)
            if token == nil {
                return false
            }
            skip_token(parser)
            if token.typ == ini_SECTION_KEY_TOKEN {
                start_mark := token.start_mark
                token := peek_token(parser)
                if token == nil {
                    return false
                }
                end_mark := token.end_mark
                skip_token(parser)
                
                parser.state = ini_PARSE_SECTION_VALUE_STATE
                tag := analyzeTag(string(token.value))
                *event = ini_event_t{
                    typ:        ini_SCALAR_EVENT,
                    start_mark: start_mark,
                    end_mark:   end_mark,
                    value:      token.value,
                    tag:        []byte(tag),
                }
                return true
            } else {
                return ini_parser_set_parser_error(parser, "did not find expected <section-key>", token.start_mark)
            }
		} else {
			return ini_parser_set_parser_error(parser, "did not find expected <section-start> or <section-key>", token.start_mark)
		}
	}
}

func ini_parser_parse_section_key(parser *ini_parser_t, event *ini_event_t, first bool) bool {
	//defer trace("ini_parser_parse_section_key")

	token := peek_token(parser)
	if token == nil {
		return false
	}
	if token.typ == ini_SECTION_KEY_TOKEN {
		skip_token(parser)
		start_mark := token.start_mark
		token := peek_token(parser)
		if token == nil {
			return false
		}
		end_mark := token.end_mark
		if token.typ == ini_SCALAR_TOKEN {
			skip_token(parser)
			parser.state = ini_PARSE_SECTION_VALUE_STATE
			tag := analyzeTag(string(token.value))
			*event = ini_event_t{
				typ:        ini_SCALAR_EVENT,
				start_mark: start_mark,
				end_mark:   end_mark,
				value:      token.value,
				tag:        []byte(tag),
				style:      ini_style_t(token.style),
			}
			return true
		} else {
			return ini_parser_set_parser_error(parser, "did not find expected <scalar>", token.start_mark)
		}
	}
	parser.state = ini_PARSE_SECTION_ENTRY_STATE
	*event = ini_event_t{
		typ:        ini_SECTION_ENTRY_EVENT,
		start_mark: token.start_mark,
		end_mark:   token.end_mark,
	}
	return true
}

// Parse the productions:
// properties           ::= (KEY = VALUE)
//
func ini_parser_parse_section_value(parser *ini_parser_t, event *ini_event_t) bool {
	//defer trace("ini_parser_parse_section_value")()

	token := peek_token(parser)
	if token == nil {
		return false
	}
 
	if token.typ == ini_SECTION_VALUE_TOKEN {
		skip_token(parser)
		start_mark := token.start_mark
		token := peek_token(parser)
		if token == nil {
			return false
		}
		end_mark := token.end_mark
		if token.typ == ini_SCALAR_TOKEN {
			skip_token(parser)
			parser.state = ini_PARSE_SECTION_KEY_STATE
			tag := analyzeTag(string(token.value))
			*event = ini_event_t{
				typ:        ini_SCALAR_EVENT,
				start_mark: start_mark,
				end_mark:   end_mark,
				value:      token.value,
				tag:        []byte(tag),
				style:      ini_style_t(token.style),
			}
			return true
		} else {
			parser.state = ini_PARSE_SECTION_KEY_STATE
			return ini_parser_process_empty_scalar(parser, event, token.start_mark)
		}
	}
	return ini_parser_set_parser_error(parser, "did not find expected <section-value>", token.start_mark)
}

func ini_parser_process_empty_scalar(parser *ini_parser_t, event *ini_event_t, mark ini_mark_t) bool {
	*event = ini_event_t{
		typ:        ini_SCALAR_EVENT,
		start_mark: mark,
		end_mark:   mark,
		value:      nil, // Empty
		tag:        []byte(ini_NULL_TAG),
		style:      ini_style_t(ini_PLAIN_SCALAR_STYLE),
	}
	return true
}
