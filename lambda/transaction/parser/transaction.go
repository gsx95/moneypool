package parser

import (
	"bytes"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"github.com/DusanKasan/parsemail"
	"github.com/ericchiang/css"
	"github.com/leekchan/accounting"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
	"transaction/data"
)

type TransactionMailParser struct {
	NameAmountRegex string
}

func NewTransactionMailParser(nameAmountRegex string) *TransactionMailParser {
	return &TransactionMailParser{NameAmountRegex: nameAmountRegex}
}

func (p *TransactionMailParser) GetTransactionInfo(email parsemail.Email) (*data.Transaction, error) {

	decodedHtml, err := b64.StdEncoding.DecodeString(email.HTMLBody)
	if err != nil {
		return nil, fmt.Errorf("Error while decoding base64 html %v", err)
	}

	rootNode, err := html.Parse(bytes.NewReader(decodedHtml))
	if err != nil {
		return nil, fmt.Errorf("Error while parsing html %v", err)
	}

	transInfo, err := p.getTransaction(rootNode)
	if err != nil {
		return nil, fmt.Errorf("Error while getting parser info %v", err)
	}

	note, err := p.getNote(rootNode)
	if err != nil {
		return nil, fmt.Errorf("Error while getting Note %v", err)
	}

	transInfo.Note = note
	return transInfo, nil

}

func (p *TransactionMailParser) getNote(html *html.Node) (string, error) {
	quoteSelector, err := css.Parse(`img[alt="quote"]`)
	if err != nil {
		return "", err
	}
	imgTags := quoteSelector.Select(html)
	if len(imgTags) < 1 {
		return "", fmt.Errorf("could not get image tag surrounding Note")
	}
	firstImg := imgTags[0]
	allTextTags, err := p.getAllSpanTexts(firstImg.Parent.Parent)
	if len(allTextTags) == 0 {
		return "", fmt.Errorf("no span text found")
	}
	if err != nil {
		return "", err
	}

	return allTextTags[0], nil
}

func (p *TransactionMailParser) getTransaction(html *html.Node) (info *data.Transaction, err error) {
	re := regexp.MustCompile(p.NameAmountRegex)
	allTexts, err := p.getAllSpanTexts(html)
	if len(allTexts) == 0 {
		return nil, fmt.Errorf("no span text found")
	}
	if err != nil {
		return nil, err
	}

	for _, line := range allTexts {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		if len(matches) < 3 {
			continue
		}

		result := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = matches[i]
			}
		}

		name := result["name"]
		amountText := result["amount"]

		base, fraction, err := p.parseAmountText(amountText)
		if err != nil {
			fmt.Println(err)
			continue
		}

		return &data.Transaction{
			Name:     name,
			Base:     base,
			Fraction: fraction,
		}, nil
	}
	return nil, errors.New("no text in html matched parser pattern")
}

func (p *TransactionMailParser) parseAmountText(amountText string) (base, fraction int, err error) {
	numReg, err := regexp.Compile("[^0-9,\\.]+")
	if err != nil {
		return
	}
	numberText := numReg.ReplaceAllString(amountText, "")
	if len(strings.TrimSpace(numberText)) == 0 {
		return 0, 0, fmt.Errorf("no amount found in amount text %s", amountText)
	}
	if !strings.ContainsAny(numberText, "0123456789") {
		return 0, 0, fmt.Errorf("no numbers found in number text %s", numberText)
	}
	curReg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		return
	}
	currency := curReg.ReplaceAllString(amountText, "")

	// the call to accounting.UnformatNumber panics if it cannot extract
	// valid locale information based on the given currency string.
	// we don't want panics, so we check the locale beforehand and return an error if necessary
	if _, localeExists := accounting.LocaleInfo[currency]; !localeExists {
		return 0, 0, fmt.Errorf("could not parse locale information from currency %s", currency)
	}

	amount := accounting.UnformatNumber(numberText, 2, currency)
	amountItems := strings.Split(amount, ".")
	baseString := strings.Replace(amountItems[0], ",", "", -1)
	fractionString := amountItems[1]
	base, err = strconv.Atoi(baseString)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse base string '%s' to int", baseString)
	}
	fraction, err = strconv.Atoi(fractionString)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse fraction string '%s' to int", fractionString)
	}
	return
}

func (p *TransactionMailParser) getAllSpanTexts(node *html.Node) ([]string, error) {
	spanSelector, err := css.Parse("p > span")
	if err != nil {
		return nil, err
	}

	var texts []string

	for _, spanNode := range spanSelector.Select(node) {
		if spanNode.FirstChild == nil {
			continue
		}
		spanText, err := p.renderNode(spanNode.FirstChild)
		if err != nil {
			continue
		}
		texts = append(texts, spanText)
	}
	return texts, nil
}

func (p *TransactionMailParser) renderNode(node *html.Node) (string, error) {
	buf := new(bytes.Buffer)
	err := html.Render(buf, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
