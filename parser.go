package main

/*
   #cgo pkg-config: purple
   #include <purple.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

var (
	emojiRegex *regexp.Regexp = regexp.MustCompile("\\[(?:\\p{Han}|[a-zA-Z])+\\]")
)

func parseMessageIntoHtml(msg []message.IMessageElement) (string, error) {
	builder := strings.Builder{}

	for _, elem := range msg {
		switch e := elem.(type) {
		case *message.FriendImageElement, *message.GroupImageElement:
			var url string
			if v, ok := e.(*message.FriendImageElement); ok {
				url = v.Url
			} else {
				url = e.(*message.GroupImageElement).Url
			}

			data, err := fetchUrl(url)
			if err != nil {
				return "", err
			}

			id := C.purple_imgstore_add_with_id(
				C.gpointer(C.CBytes(data)), C.ulong(len(data)), nil)
			builder.WriteString(fmt.Sprintf("<img id=\"%d\">", id))
		case *message.TextElement:
			builder.WriteString(html.EscapeString(e.Content))
		case *message.FaceElement:
			builder.WriteString(fmt.Sprintf("[%s]", e.Name))
		case *message.GroupFileElement:
			builder.WriteString(fmt.Sprintf("Upload %s(%s)", e.Name, sizeToString(e.Size)))
		default:
			debugWarning("parseMessageIntoHtml: ignoring %v\n", elem)
		}
	}

	return builder.String(), nil
}

type MessagingContext struct {
	client  *client.QQClient
	isGroup bool
	id      int64
}

func parseHtmlIntoMessage(ctx MessagingContext, msg string) ([]message.IMessageElement, error) {
	elems := make([]message.IMessageElement, 0)
	doc, err := html.Parse(strings.NewReader(msg))
	if err != nil {
		return nil, err
	}

	var body *html.Node
	for n := doc.FirstChild.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.DataAtom == atom.Body {
			body = n
			break
		}
	}
	if body == nil {
		return nil, fmt.Errorf("Could not find <body>")
	}

	for n := body.FirstChild; n != nil; n = n.NextSibling {
		switch n.Type {
		case html.ElementNode:
			break
		case html.TextNode:
			elems = append(elems, parseText(n.Data)...)
			continue
		default:
			debugWarning("parseHtmlIntoMessage: ignoring %s(%d)\n",
				n.Data, n.Type)
			continue
		}

		switch n.DataAtom {
		case atom.Img:
			id, err := strconv.Atoi(*findAttr(n.Attr, "id"))
			if err != nil {
				return nil, err
			}

			img := C.purple_imgstore_find_by_id(C.int(id))
			data := C.GoBytes(
				unsafe.Pointer(C.purple_imgstore_get_data(img)),
				C.int(C.purple_imgstore_get_size(img)),
			)

			var imgelem message.IMessageElement
			if ctx.isGroup {
				imgelem, err = ctx.client.UploadGroupImage(
					ctx.id, bytes.NewReader(data))
			} else {
				imgelem, err = ctx.client.UploadPrivateImage(
					ctx.id, bytes.NewReader(data))
			}

			if err != nil {
				return nil, err
			}

			elems = append(elems, imgelem)
		case atom.Br, atom.Hr:
			elems = append(elems, message.NewText("\n"))
		default:
			debugWarning("parseHtmlIntoMessage: ignoring %s(%d)\n",
				n.Data, n.Type)
		}
	}

	return elems, nil
}

func parseText(str string) []message.IMessageElement {
	i := 0
	elems := make([]message.IMessageElement, 0)

	for _, idx := range emojiRegex.FindAllIndex([]byte(str), -1) {
		id := getFaceId(str[idx[0]+1: idx[1]-1])
		if id == -1 {
			continue
		}

		if idx[0] > i {
			elems = append(elems, message.NewText(str[i:idx[0]]))
		}

		i = idx[1]
		elems = append(elems, message.NewFace(id))
	}

	if i != len(str) {
		elems = append(elems, message.NewText(str[i:]))
	}

	return elems
}

func findAttr(arr []html.Attribute, key string) *string {
	for _, attr := range arr {
		if strings.EqualFold(attr.Key, key) {
			return &attr.Val
		}
	}

	return nil
}
