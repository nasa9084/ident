package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	openapi "github.com/nasa9084/go-openapi"
)

type options struct {
	File string `short:"f" long:"file" description:"spec file"`
}

func main() { os.Exit(_main()) }

func _main() int {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		log.Print(err)
		return 1
	}
	spec, err := openapi.Load(opts.File)
	if err != nil {
		log.Print(err)
		return 1
	}
	if err := generateRequests(spec); err != nil {
		log.Print(err)
		return 1
	}

	if err := generateResponses(spec); err != nil {
		log.Print(err)
		return 1
	}
	if err := generateHandlers(spec); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func generateRequests(spec *openapi.Document) error {
	buf := bytes.Buffer{}
	buf.WriteString("package input")
	buf.WriteString("\nimport (")
	buf.WriteString("\n\"errors\"")
	buf.WriteString("\n\"unicode\"")
	buf.WriteString("\n)")
	buf.WriteString("\nvar _ = unicode.UpperCase")
	if err := generateRequestInterface(&buf); err != nil {
		return err
	}
	for _, pathItem := range spec.Paths {
		if err := generateRequestsForPathItem(&buf, pathItem); err != nil {
			return err
		}
	}
	buf.WriteString("\n")
	return writeTo(buf.Bytes(), "usecase/input/request_gen.go")
}

func generateRequestInterface(buf *bytes.Buffer) error {
	buf.WriteString("\n\ntype Request interface{")
	buf.WriteString("\nValidate() error")
	buf.WriteString("\n}")
	buf.WriteString("\n\ntype SessionRequest interface {")
	buf.WriteString("\nRequest")
	buf.WriteString("\nSetSessionID(string)")
	buf.WriteString("\n}")
	buf.WriteString("\n\ntype PathArgsRequest interface {")
	buf.WriteString("\nRequest")
	buf.WriteString("\nSetPathArgs(map[string]string)")
	buf.WriteString("\n}")
	return nil
}

func generateRequestsForPathItem(buf *bytes.Buffer, pathItem *openapi.PathItem) error {
	if pathItem.Get != nil {
		if err := generateRequest(buf, pathItem.Get); err != nil {
			return err
		}
	}
	if pathItem.Post != nil {
		if err := generateRequest(buf, pathItem.Post); err != nil {
			return err
		}
	}
	if pathItem.Put != nil {
		if err := generateRequest(buf, pathItem.Put); err != nil {
			return err
		}
	}
	return nil
}

func generateRequest(buf *bytes.Buffer, op *openapi.Operation) error {
	if op.RequestBody == nil && op.Parameters == nil && op.Security == nil {
		return nil
	}
	buf.WriteString("\n\ntype ")
	buf.WriteString(op.OperationID)
	buf.WriteString("Request struct {")
	if op.RequestBody != nil {
		for k, v := range op.RequestBody.Content["application/json"].Schema.Properties {
			buf.WriteString("\n")
			buf.WriteString(v.Title)
			buf.WriteString("\t")
			if v.Type == "string" {
				buf.WriteString("string")
			} else {
				return errors.New("unknown type")
			}
			buf.WriteString("\t `json:\"")
			buf.WriteString(k)
			buf.WriteString("\"`")
		}
	}
	var isSessionRequest bool
	if op.Security != nil {
		for _, security := range *op.Security {
			if _, ok := security["sessionId"]; ok {
				buf.WriteString("\n\nSessionID string `json:\"-\"`")
				isSessionRequest = true
			}
		}
	}
	var isPathArgsRequest bool
	pathArgTitles := map[string]*openapi.Schema{}
	for _, param := range op.Parameters {
		if param.In == "path" {
			buf.WriteString("\n\n")
			buf.WriteString(param.Schema.Title)
			buf.WriteString(" ")
			buf.WriteString(param.Schema.Type)
			buf.WriteString(" `json:\"-\"`")
			isPathArgsRequest = true
			pathArgTitles[param.Name] = param.Schema
		}
	}
	buf.WriteString("\n}")

	buf.WriteString("\nfunc (r ")
	buf.WriteString(op.OperationID)
	buf.WriteString("Request) Validate() error {")
	formatDigitMap := map[string]string{}
	buf.WriteString("\nswitch {")
	for n, s := range pathArgTitles {
		buf.WriteString("\ncase r.")
		buf.WriteString(s.Title)
		buf.WriteString(" == ")
		if s.Type == "string" {
			buf.WriteString(strconv.Quote(""))
		} else {
			buf.WriteString("nil")
		}
		buf.WriteString(":")
		buf.WriteString("\nreturn errors.New(")
		buf.WriteString(strconv.Quote(fmt.Sprintf("%s is required", n)))
		buf.WriteString(")")
	}
	if isSessionRequest {
		buf.WriteString("\ncase r.SessionID == \"\":")
		buf.WriteString("\nreturn errors.New(")
		buf.WriteString(strconv.Quote("authorization header is required"))
		buf.WriteString(")")
	}
	if op.RequestBody != nil {
		for _, required := range op.RequestBody.Content["application/json"].Schema.Required {
			buf.WriteString("\ncase r.")
			buf.WriteString(op.RequestBody.Content["application/json"].Schema.Properties[required].Title)
			buf.WriteString(` == `)
			if op.RequestBody.Content["application/json"].Schema.Properties[required].Type == "string" {
				buf.WriteString(`""`)
			} else {
				buf.WriteString("nil")
			}
			buf.WriteString(":")
			buf.WriteString("\nreturn errors.New(\"")
			buf.WriteString(required)
			buf.WriteString(" is required \")")
		}
		for p, s := range op.RequestBody.Content["application/json"].Schema.Properties {
			if s.MaxLength != 0 && s.MaxLength == s.MinLength {
				buf.WriteString("\ncase len(r.")
				buf.WriteString(s.Title)
				buf.WriteString(") != ")
				buf.WriteString(strconv.Itoa(s.MaxLength))
				buf.WriteString(":")
				buf.WriteString("\nreturn errors.New(\"length of ")
				buf.WriteString(p)
				buf.WriteString(" is not valid\")")
			} else {
				if s.MaxLength != 0 {
					buf.WriteString("\ncase len(r.")
					buf.WriteString(s.Title)
					buf.WriteString(") > ")
					buf.WriteString(strconv.Itoa(s.MaxLength))
					buf.WriteString(":")
					buf.WriteString("\nreturn errors.New(\"length of ")
					buf.WriteString(p)
					buf.WriteString(" is over")
				}
				if s.MinLength != 0 {
					buf.WriteString("\ncase len(r.")
					buf.WriteString(s.Title)
					buf.WriteString(") < ")
					buf.WriteString(strconv.Itoa(s.MinLength))
					buf.WriteString(":")
					buf.WriteString("\nreturn errors.New(\"length of ")
					buf.WriteString(p)
					buf.WriteString(" is less")
				}
			}
			if s.Format == "digit" {
				formatDigitMap[p] = s.Title
			}
		}
	}

	buf.WriteString("\n}")
	for p, t := range formatDigitMap {
		buf.WriteString("\nfor _, r := range r.")
		buf.WriteString(t)
		buf.WriteString(" {")
		buf.WriteString("\nif !unicode.IsDigit(r) {")
		buf.WriteString("\nreturn errors.New(")
		buf.WriteString(strconv.Quote(fmt.Sprintf("%s must be digit", p)))
		buf.WriteString(")")
		buf.WriteString("\n}")
		buf.WriteString("\n}")
	}
	buf.WriteString("\nreturn nil")
	buf.WriteString("\n}")

	if !isSessionRequest {
		return nil
	}
	buf.WriteString("\n\nfunc (r ")
	buf.WriteString(op.OperationID)
	buf.WriteString("Request) SetSessionID(sessid string) {")
	buf.WriteString("\nr.SessionID = sessid")
	buf.WriteString("\n}")

	if isPathArgsRequest {
		buf.WriteString("\n\nfunc (r ")
		buf.WriteString(op.OperationID)
		buf.WriteString("Request) SetPathArgs(args map[string]string) {")
		for _, param := range op.Parameters {
			if param.In != "path" {
				continue
			}
			buf.WriteString("\nr.")
			buf.WriteString(param.Schema.Title)
			buf.WriteString(" = ")
			buf.WriteString(param.Name)
		}
		buf.WriteString("\n}")
	}

	return nil
}

func generateResponses(spec *openapi.Document) error {
	var buf bytes.Buffer
	buf.WriteString("package output")
	buf.WriteString("\nimport (")
	writeImport(&buf, "encoding/json")
	writeImport(&buf, "net/http")
	buf.WriteString("\n")
	writeImport(&buf, "github.com/lestrrat-go/bufferpool")
	buf.WriteString("\n)")
	if err := generateResponseInterface(&buf, spec.Components.Responses); err != nil {
		return err
	}
	for _, pathItem := range spec.Paths {
		if err := generateResponseForPathItem(&buf, pathItem); err != nil {
			return err
		}
	}
	return writeTo(buf.Bytes(), "usecase/output/response_gen.go")
}

func generateResponseInterface(buf *bytes.Buffer, responses openapi.Responses) error {
	buf.WriteString(fmt.Sprintf("\n\nvar okBody = map[string]string{%s: %s}", strconv.Quote("message"), strconv.Quote("ok")))
	buf.WriteString("\n\ntype Response interface {")
	buf.WriteString("\nRender(http.ResponseWriter)")
	buf.WriteString("\n}")
	for n, resp := range responses {
		buf.WriteString("\n\ntype ")
		buf.WriteString(n)
		buf.WriteString(" struct {")
		for pn, p := range resp.Content["application/json"].Schema.Properties {
			buf.WriteString("\n")
			buf.WriteString(p.Title)
			buf.WriteString("\t")
			buf.WriteString(p.Type)
			buf.WriteString("`json:\"")
			buf.WriteString(pn)
			buf.WriteString("\"`")
		}
		buf.WriteString("\n}")
	}
	buf.WriteString("\n\nfunc renderJSON(w http.ResponseWriter, status int, v interface{}) {")
	buf.WriteString("\nif v == nil {")
	buf.WriteString("\nje := jsonErr{")
	buf.WriteString("\nMessage: \"nil response\",")
	buf.WriteString("\nError: http.StatusText(http.StatusInternalServerError),")
	buf.WriteString("\n}")
	buf.WriteString("\nrenderJSON(w, status, je)")
	buf.WriteString("\nreturn")
	buf.WriteString("\n}")
	buf.WriteString("\nif err, ok := v.(error); ok {")
	buf.WriteString("\nje := jsonErr {")
	buf.WriteString("\nMessage: err.Error(),")
	buf.WriteString("\nError: http.StatusText(status),")
	buf.WriteString("\n}")
	buf.WriteString("\nrenderJSON(w, status, je)")
	buf.WriteString("\nreturn")
	buf.WriteString("\n}")
	buf.WriteString("\nbuf := bufferpool.Get()")
	buf.WriteString("\ndefer bufferpool.Release(buf)")
	buf.WriteString("\nif err := json.NewEncoder(buf).Encode(v); err != nil {")
	buf.WriteString("\nrenderJSON(w, http.StatusInternalServerError, err)")
	buf.WriteString("\nreturn")
	buf.WriteString("\n}")
	buf.WriteString("\nw.Header().Set(\"Content-Type\", \"application/json\")")
	buf.WriteString("\nw.WriteHeader(status)")
	buf.WriteString("\nbuf.WriteTo(w)")
	buf.WriteString("\n}")
	buf.WriteString("\n\nfunc renderPEM(w http.ResponseWriter, status int, pem []byte) {")
	buf.WriteString("\nw.Header().Set(\"Content-Type\", \"application/x-pem-file\")")
	buf.WriteString("\nw.WriteHeader(status)")
	buf.WriteString("\nw.Write(pem)")
	buf.WriteString("\n}")
	buf.WriteString("\n\nfunc renderPNG(w http.ResponseWriter, status int, png []byte) {")
	buf.WriteString("\nw.Header().Set(\"Content-Type\", \"image/png\")")
	buf.WriteString("\nw.WriteHeader(status)")
	buf.WriteString("\nw.Write(png)")
	buf.WriteString("\n}")

	buf.WriteString("\n\nfunc renderJSONWithSessionID(w http.ResponseWriter, status int, err error, sessid string) {")
	buf.WriteString("\nif err != nil {")
	buf.WriteString("\nrenderJSON(w, status, err)")
	buf.WriteString("\nreturn")
	buf.WriteString("\n}")
	buf.WriteString("\nw.Header().Set(\"X-SESSION-ID\", sessid)")
	buf.WriteString("\nrenderJSON(w, status, map[string]string{\"message\":\"ok\"})")
	buf.WriteString("\n}")

	return nil
}

func generateResponseForPathItem(buf *bytes.Buffer, pathItem *openapi.PathItem) error {
	if pathItem.Get != nil {
		if err := generateResponse(buf, pathItem.Get); err != nil {
			return err
		}
	}
	if pathItem.Post != nil {
		if err := generateResponse(buf, pathItem.Post); err != nil {
			return err
		}
	}
	if pathItem.Put != nil {
		if err := generateResponse(buf, pathItem.Put); err != nil {
			return err
		}
	}
	return nil
}

func generateResponse(buf *bytes.Buffer, op *openapi.Operation) error {
	buf.WriteString(fmt.Sprintf("\n\ntype %sResponse struct {", op.OperationID))
	buf.WriteString("\nStatus int")
	buf.WriteString("\nErr error")
	resp, ok := op.Responses["200"]
	if !ok {
		resp, ok = op.Responses["201"]
	}
	if ok {
		buf.WriteString("\n")
		for _, mime := range resp.Content {
			if mime.Schema.Type == "object" {
				for _, p := range mime.Schema.Properties {
					buf.WriteString(fmt.Sprintf("\n%s %s", p.Title, p.Type))
				}
			} else if mime.Schema.Type == "string" {
				if mime.Schema.Title != "" {
					buf.WriteString(fmt.Sprintf("\n%s ", mime.Schema.Title))
				} else {
					buf.WriteString("\nBody ")
				}
				switch mime.Schema.Format {
				case "binary":
					buf.WriteString("[]byte")
				default:
					buf.WriteString("string")
				}
			}
		}
	}
	var returnSessionID bool
	if resp.Headers != nil {
		if _, ok := resp.Headers["X-SESSION-ID"]; ok {
			buf.WriteString("\n\nSessionID string")
			returnSessionID = true
		}
	}
	buf.WriteString("\n}")

	buf.WriteString(fmt.Sprintf("\n\nfunc (resp %sResponse) Render(w http.ResponseWriter) {", op.OperationID))
	buf.WriteString("\nif resp.Err != nil {")
	buf.WriteString("\nrenderJSON(w, resp.Status, resp.Err)")
	buf.WriteString("\nreturn")
	buf.WriteString("\n}")
	if returnSessionID {
		buf.WriteString("\nrenderJSONWithSessionID(w, resp.Status, resp.Err, resp.SessionID)")
	} else {
		buf.WriteString("\nrenderJSON(w, resp.Status, okBody)")
	}
	buf.WriteString("\n}")
	return nil
}

func generateHandlers(spec *openapi.Document) error {
	var buf bytes.Buffer
	buf.WriteString("package ident")
	buf.WriteString("\n\nimport (")
	writeImport(&buf, "encoding/json")
	writeImport(&buf, "fmt")
	writeImport(&buf, "net/http")
	writeImport(&buf, "strings")
	buf.WriteString("\n")
	writeImport(&buf, "github.com/gorilla/mux")
	writeImport(&buf, "github.com/lestrrat-go/bufferpool")
	writeImport(&buf, "github.com/nasa9084/ident/infra")
	writeImport(&buf, "github.com/nasa9084/ident/usecase")
	writeImport(&buf, "github.com/nasa9084/ident/usecase/input")
	writeImport(&buf, "github.com/pkg/errors")
	buf.WriteString("\n)")
	if err := generateHandlerHelper(&buf); err != nil {
		return err
	}
	if err := generateErrHandler(&buf); err != nil {
		return err
	}
	for _, pathItem := range spec.Paths {
		if err := generateHandlerForPathItem(&buf, pathItem); err != nil {
			return err
		}
	}
	return writeTo(buf.Bytes(), "handler_gen.go")
}

func writeImport(buf *bytes.Buffer, pkg string) {
	buf.WriteString(fmt.Sprintf("\n%s", strconv.Quote(pkg)))
}

func generateHandlerHelper(buf *bytes.Buffer) error {
	buf.WriteString("\n\nfunc parseRequest(r *http.Request, dest input.Request) error {")
	buf.WriteString("\nif r.Method != http.MethodGet {")
	buf.WriteString("\nif err := json.NewDecoder(r.Body).Decode(dest); err != nil {")
	buf.WriteString(fmt.Sprintf("\nreturn errors.Wrap(err, %s)", strconv.Quote("parsing request body")))
	buf.WriteString("\n}")
	buf.WriteString("\n}")
	buf.WriteString("\nif sessReq, ok := dest.(input.SessionRequest); ok {")
	buf.WriteString(fmt.Sprintf("\nauthorization := r.Header.Get(%s)", strconv.Quote("Authorization")))
	buf.WriteString("\nif strings.Contains(authorization, ` `) {")
	buf.WriteString(fmt.Sprintf("\nreturn errors.New(%s)", strconv.Quote("authorization header invalid")))
	buf.WriteString("\n}")
	buf.WriteString("\nsessReq.SetSessionID(authorization)")
	buf.WriteString("\n}")
	buf.WriteString("\nif arReq, ok := dest.(input.PathArgsRequest); ok {")
	buf.WriteString("\narReq.SetPathArgs(mux.Vars(r))")
	buf.WriteString("\n}")
	buf.WriteString("\nreturn dest.Validate()")
	buf.WriteString("\n}")

	buf.WriteString("\n\nfunc renderErr(w http.ResponseWriter, err error) {")
	buf.WriteString("\nbuf := bufferpool.Get()")
	buf.WriteString("\ndefer bufferpool.Release(buf)")
	buf.WriteString("\n\nv := map[string]string{")
	buf.WriteString(fmt.Sprintf("\n\n%s: http.StatusText(http.StatusBadRequest),", strconv.Quote("error")))
	buf.WriteString(fmt.Sprintf("\n\n%s: err.Error(),", strconv.Quote("message")))
	buf.WriteString("\n}")
	buf.WriteString("\njson.NewEncoder(buf).Encode(v)")
	buf.WriteString(fmt.Sprintf("\nw.Header().Set(%s, %s)", strconv.Quote("Content-Type"), strconv.Quote("application/json")))
	buf.WriteString("\nw.WriteHeader(http.StatusBadRequest)")
	buf.WriteString("\nbuf.WriteTo(w)")
	buf.WriteString("\n}")

	return nil
}

func generateErrHandler(buf *bytes.Buffer) error {
	buf.WriteString("\n\nfunc NotFoundHandler(w http.ResponseWriter, r *http.Request) {")
	buf.WriteString("\nbuf := bufferpool.Get()")
	buf.WriteString("\ndefer bufferpool.Release(buf)")
	buf.WriteString("\n\nv := map[string]string{")
	buf.WriteString(fmt.Sprintf("\n%s: http.StatusText(http.StatusNotFound),", strconv.Quote("error")))
	buf.WriteString(fmt.Sprintf("\n%s: %s,", strconv.Quote("message"), strconv.Quote("endpoint not found")))
	buf.WriteString("\n}")
	buf.WriteString("\njson.NewEncoder(buf).Encode(v)")
	buf.WriteString(fmt.Sprintf("\nw.Header().Set(%s, %s)", strconv.Quote("Content-Type"), strconv.Quote("application/json")))
	buf.WriteString("\nw.WriteHeader(http.StatusNotFound)")
	buf.WriteString("\nbuf.WriteTo(w)")
	buf.WriteString("\n}")

	buf.WriteString("\n\nfunc MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {")
	buf.WriteString("\nbuf := bufferpool.Get()")
	buf.WriteString("\ndefer bufferpool.Release(buf)")
	buf.WriteString("\n\nv := map[string]string{")
	buf.WriteString(fmt.Sprintf("\n%s: http.StatusText(http.StatusMethodNotAllowed),", strconv.Quote("error")))
	buf.WriteString(fmt.Sprintf("\n%s: fmt.Sprintf(%s, r.Method),", strconv.Quote("message"), strconv.Quote("method %s is not allowed")))
	buf.WriteString("\n}")
	buf.WriteString("\njson.NewEncoder(buf).Encode(v)")
	buf.WriteString(fmt.Sprintf("\nw.Header().Set(%s, %s)", strconv.Quote("Content-Type"), strconv.Quote("application/json")))
	buf.WriteString("\nw.WriteHeader(http.StatusMethodNotAllowed)")
	buf.WriteString("\nbuf.WriteTo(w)")
	buf.WriteString("\n}")
	return nil
}

func generateHandlerForPathItem(buf *bytes.Buffer, pathItem *openapi.PathItem) error {
	if pathItem.Get != nil {
		if err := generateHandler(buf, pathItem.Get); err != nil {
			return err
		}
	}
	if pathItem.Post != nil {
		if err := generateHandler(buf, pathItem.Post); err != nil {
			return err
		}
	}
	if pathItem.Put != nil {
		if err := generateHandler(buf, pathItem.Put); err != nil {
			return err
		}
	}

	return nil
}

func generateHandler(buf *bytes.Buffer, op *openapi.Operation) error {
	buf.WriteString(fmt.Sprintf("\n\nfunc %sHandler(env *infra.Environment) http.HandlerFunc {", op.OperationID))
	buf.WriteString("\nreturn func(w http.ResponseWriter, r *http.Request) {")
	if op.RequestBody != nil || op.Parameters != nil || op.Security != nil {
		buf.WriteString(fmt.Sprintf("\nvar req input.%sRequest", op.OperationID))
		buf.WriteString("\nif err := parseRequest(r, &req); err != nil {")
		buf.WriteString("\nrenderErr(w, err)")
		buf.WriteString("\nreturn")
		buf.WriteString("\n}")
	}
	buf.WriteString(fmt.Sprintf("\nusecase.%s(r.Context(), ", op.OperationID))
	if op.RequestBody != nil || op.Parameters != nil || op.Security != nil {
		buf.WriteString("req, ")
	}
	buf.WriteString("env).Render(w)")
	buf.WriteString("\n}")
	buf.WriteString("\n}")
	return nil
}

func writeTo(src []byte, out string) error {
	f, err := os.OpenFile(out, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write to file
	formatted, err := format.Source(src)
	if err != nil {
		return err
	}

	// add generated code comment
	formatted = append(
		[]byte("// Code generated by genHandler. DO NOT EDIT.\n\n"),
		formatted...,
	)
	_, err = f.Write(formatted)

	return err
}
