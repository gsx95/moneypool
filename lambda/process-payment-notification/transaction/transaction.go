package transaction

import (
	"bytes"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/ericchiang/css"
	"github.com/leekchan/accounting"
	"golang.org/x/net/html"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Info struct {
	Name     string
	Base     int
	Fraction int
	Note     string
}

var (
	NameAmountRegex = os.Getenv("NameAmountRegex")
)

func GetTransactionInfo(email parsemail.Email) (*Info, error) {

	decodedHtml, err := b64.StdEncoding.DecodeString(email.HTMLBody)
	if err != nil {
		return nil, fmt.Errorf("Error while decoding base64 html %v\n", err)
	}

	rootNode, err := html.Parse(bytes.NewReader(decodedHtml))
	if err != nil {
		return nil, fmt.Errorf("Error while parsing html %v\n", err)
	}

	transInfo, err := getTransactionInfo(rootNode)
	if err != nil {
		return nil, fmt.Errorf("Error while getting transaction info %v\n", err)
	}

	note, err := getNote(rootNode)
	if err != nil {
		return nil, fmt.Errorf("Error while getting Note %v\n", err)
	}

	transInfo.Note = note
	return transInfo, nil

}

func getNote(html *html.Node) (string, error) {
	quoteSelector, err := css.Parse(`img[alt="quote"]`)
	if err != nil {
		return "", err
	}
	imgTags := quoteSelector.Select(html)
	if len(imgTags) < 1 {
		return "", fmt.Errorf("could not get image tag surrounding Note")
	}
	firstImg := imgTags[0]
	allTextTags, err := getAllSpanTexts(firstImg.Parent.Parent)
	if err != nil {
		return "", err
	}

	return allTextTags[0], nil
}

func getTransactionInfo(html *html.Node) (info *Info, err error) {
	re := regexp.MustCompile(NameAmountRegex)
	allTexts, err := getAllSpanTexts(html)
	if err != nil {
		return nil, err
	}

	for _, line := range allTexts {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			fmt.Println("line not matching")
			continue
		}
		if len(matches) != 3 {
			fmt.Println("regex matched, but wrong count of subexpressions found")
			continue
		}

		name := matches[1]
		amountText := matches[2]

		base, fraction, err := parseAmountText(amountText)
		if err != nil {
			fmt.Println("regex matched, but wrong count of subexpressions found")
			continue
		}

		return &Info{
			Name:     name,
			Base:     base,
			Fraction: fraction,
		}, nil
	}
	return nil, errors.New("no text in html matched transaction pattern")
}

func parseAmountText(amountText string) (base, fraction int, err error) {
	numReg, err := regexp.Compile("[^0-9,\\.]+")
	if err != nil {
		return
	}
	numberText := numReg.ReplaceAllString(amountText, "")
	curReg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		return
	}
	currency := curReg.ReplaceAllString(amountText, "")
	amount := accounting.UnformatNumber(numberText, 2, currency)
	amountItems := strings.Split(amount, ".")
	baseString := strings.Replace(amountItems[0], ",", "", -1)
	fractionString := amountItems[1]
	base, err = strconv.Atoi(baseString)
	if err != nil {
		return 0, 0, nil
	}
	fraction, err = strconv.Atoi(fractionString)
	if err != nil {
		return 0, 0, nil
	}
	return
}

func getAllSpanTexts(node *html.Node) ([]string, error) {
	spanSelector, err := css.Parse("p > span")
	if err != nil {
		return nil, err
	}

	var texts []string

	for _, spanNode := range spanSelector.Select(node) {
		spanText, err := renderNode(spanNode.FirstChild)
		if err != nil {
			continue
		}
		texts = append(texts, spanText)
	}
	return texts, nil
}

func renderNode(node *html.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := html.Render(buf, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
