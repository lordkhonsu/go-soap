package dom

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	xPathOutputDebug        = false
	xPathIndexCondition     = regexp.MustCompile(`^[0-9]+$`)
	xPathAttributeCondition = regexp.MustCompile(`^@(.+?)(='(.+)')?$`)
	xPathLastCondition      = regexp.MustCompile(`^last\(\)(-([0-9]+))?$`)
	xPathEscaper            = strings.NewReplacer(
		`'`, `%27`,
		`/`, `%2F`,
		`[`, `%5B`,
		`]`, `%5D`,
	)
)

// SetXPathDebugOutput controls if the xpath system should output debug messages
func SetXPathDebugOutput(printOutput bool) {
	xPathOutputDebug = true
}

// GetXPath builds a XPath to this node
func (n *Node) GetXPath() string {
	if n.Parent != nil {
		return n.Parent.GetXPath() + "/" + n.Parent.GetXPathRel(n)
	}
	return "/" + n.Name
}

// GetXPathRel returns a XPath relation to the specified child node
func (n *Node) GetXPathRel(rel *Node) string {
	index := 0
	for _, child := range n.Children.All() {
		if child == rel {
			return fmt.Sprintf("%s[%d]", rel.Name, index+1)
		}
		if child.Name == rel.Name {
			index++
		}
	}
	panic(fmt.Errorf("GetXPath(): node not found in children of defined parent"))
}

// GetXPathName returns a XPath relation to a possible new child node with a given name
func (n *Node) GetXPathName(name string) string {
	index := 0
	for _, child := range n.Children.All() {
		if child.Name == name {
			index++
		}
	}
	return fmt.Sprintf("%s[%d]", name, index+1)
}

// XPath resolves the given XPath to a list of Nodes
func (n *Node) XPath(xpath string, arguments ...interface{}) *NodeList {
	if !n.Exists {
		return &NodeList{}
	}

	// transform arguments into xpath
	if len(arguments) > 0 {
		xpath = n.xPathPrintf(xpath, arguments...)
	}

	// get next xpath token
	next, xpath := n.nextXPathToken(xpath)

	xPathDebug("[xpath] self=%s, next=%s, xpath=%s\n", n.Name, next, xpath)

	// we are done
	if next == "" {
		return &NodeList{nodes: []*Node{n}}
	}

	// goto root
	if next == "/" {
		return n.Document.XPath(xpath)
	}

	// complicated
	if next == "//" {
		panic(fmt.Errorf("XPath(): // is not yet supported"))
	}

	// why?
	if next == "." || (next == n.Name && n.Parent == nil) {
		return n.XPath(xpath)
	}

	// why??
	if next == ".." {
		if n.Parent != nil {
			return n.Parent.XPath(xpath)
		}
		panic(fmt.Errorf("XPath(): element has no parent"))
	}

	name, condition := n.splitXPathToken(next)

	// name comparison
	result := &NodeList{}
	nodes := n.Children.All()
	count := len(nodes)

	namedIndex := map[string]int{}

	for _, child := range nodes {
		// compare name
		if n.xPathEscape(child.Name) == name || name == "*" {
			if _, exists := namedIndex[child.Name]; !exists {
				namedIndex[child.Name] = 0
			}
			namedIndex[child.Name]++

			// condition test
			if !n.checkXPathCondition(condition, child, namedIndex[child.Name], count) {
				continue
			}

			// append result
			childResult := child.XPath(xpath)
			result.AppendList(childResult)
		}
	}

	return result
}

func (n *Node) xPathEscape(in string) string {
	return xPathEscaper.Replace(in)
}

func (n *Node) xPathPrintf(xpath string, arguments ...interface{}) string {
	list := make([]interface{}, len(arguments))
	// escape arguments
	for i, arg := range arguments {
		list[i] = n.xPathEscape(fmt.Sprint(arg))
	}
	// printf
	return fmt.Sprintf(xpath, list...)
}

func (n *Node) checkXPathCondition(condition string, node *Node, index int, count int) bool {
	if condition == "" {
		return true
	}

	// simple index (per W3C standard, counting begins at 1)
	if xPathIndexCondition.MatchString(condition) {
		expect, _ := strconv.ParseInt(condition, 10, 64)
		xPathDebug("[xpath] self=%s, xPathIndexCondition: index=%d\n", node.Name, expect)
		return expect == int64(index)
	}

	// attribute check
	if matches := xPathAttributeCondition.FindStringSubmatch(condition); len(matches) > 0 {
		if matches[1] == "*" {
			xPathDebug("[xpath] self=%s, xPathAttributeCondition: name=%s, check=any\n", node.Name, matches[1])
			return len(node.Attributes) > 0
		}
		if matches[3] == "" {
			xPathDebug("[xpath] self=%s, xPathAttributeCondition: name=%s, check=exists\n", node.Name, matches[1])
			_, exists := node.GetAttributeX(matches[1])
			return exists
		}
		xPathDebug("[xpath] self=%s, xPathAttributeCondition: name=%s, check=value, value=%s\n", node.Name, matches[1], matches[3])
		return n.xPathEscape(node.GetAttributeXValue(matches[1])) == matches[3]
	}

	// conditional position (last() and last()-N)
	if matches := xPathLastCondition.FindStringSubmatch(condition); len(matches) > 0 {
		shift := int64(0)
		if matches[2] != "" {
			shift, _ = strconv.ParseInt(matches[2], 10, 64)
		}
		xPathDebug("[xpath] self=%s, xPathLastCondition: shift=%d\n", node.Name, shift)
		return index == count-int(shift)
	}

	panic(fmt.Errorf("XPath(): condition not yet supported: %s", condition))
}

func (n *Node) nextXPathToken(xpath string) (string, string) {
	// we are done here
	if len(xpath) == 0 {
		return "", ""
	}

	// `//` -> deep search
	if len(xpath) > 2 && xpath[0:2] == "//" {
		return "//", xpath[2:]
	}

	// `/` -> jump to document root
	if xpath[0:1] == "/" {
		return "/", xpath[1:]
	}

	// find next splitter
	index := strings.Index(xpath, "/")
	if index == -1 {
		// no more splitters -> last token
		return xpath, ""
	}

	// splitter found -> split
	return xpath[:index], xpath[index+1:]
}

func (n *Node) splitXPathToken(xpathToken string) (string, string) {
	index := strings.Index(xpathToken, "[")

	// no conditions set
	if index == -1 {
		return xpathToken, ""
	}

	return xpathToken[:index], xpathToken[index+1 : len(xpathToken)-1]
}

func xPathDebug(format string, arguments ...interface{}) {
	if !xPathOutputDebug {
		return
	}
	fmt.Printf(format, arguments...)
}
