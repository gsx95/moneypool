package parser

import (
	"bytes"
	b64 "encoding/base64"
	"errors"
	"github.com/DusanKasan/parsemail"
	"io/ioutil"
	"os"
	"testing"
	"text/template"
	"transaction/data"
)

const nameAmountRegex = "(?P<name>(.+)) hat Ihnen (?P<amount>(.+)) gesendet"

var mailTemplate *template.Template

type TransactionTest struct {
	name        string
	inputMail   parsemail.Email
	expectedOut *data.Transaction
	expectError error
}

func TestMain(m *testing.M) {
	mailTemplateText, err := ioutil.ReadFile("tests/template.mail")
	if err != nil {
		panic(err)
	}
	mailTemplate, err = template.New("mailTest").Parse(string(mailTemplateText))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestGetTransactionInfoNames(t *testing.T) {
	testTable := []TransactionTest{
		{
			"valid_name_two_words",
			getEmail(mailTemplate, "tests/name/valid_name_two_words.html", true),
			&data.Transaction{Name: "Sender Person"},
			nil,
		},
		{
			"valid_name_multi_word",
			getEmail(mailTemplate, "tests/name/valid_name_multi_word.html", true),
			&data.Transaction{Name: "Sender MiddlenameOne MiddlenameTwo Person"},
			nil,
		},
		{
			"valid_name_special_chars",
			getEmail(mailTemplate, "tests/name/valid_name_special_chars.html", true),
			&data.Transaction{Name: "Björk Guðmundsdóttirñ"},
			nil,
		},
		{
			"valid_name_special_chars_2",
			getEmail(mailTemplate, "tests/name/valid_name_special_chars_2.html", true),
			&data.Transaction{Name: "秀英張"},
			nil,
		},
	}
	for _, test := range testTable {
		parser := NewTransactionMailParser(nameAmountRegex)
		output, err := parser.GetTransactionInfo(test.inputMail)
		if !compareErrors(err, test.expectError) {
			t.Fatalf("GetTransactionInfo(%s) returned error %v, but should return with error %v", test.name, err, test.expectError)
		}
		if test.expectedOut == nil {
			return
		}
		if output.Name != test.expectedOut.Name {
			t.Fatalf("GetTransactionInfo(%s) returned name %v, but should return %v", test.name, output.Name, test.expectedOut.Name)
		}
	}
}

func TestGetTransactionInfoAmounts(t *testing.T) {
	testTable := []TransactionTest{
		{
			"valid_amount_1,99EUR",
			getEmail(mailTemplate, "tests/amount/valid_amount_1,99EUR.html", true),
			&data.Transaction{Base: 1, Fraction: 99},
			nil,
		},
		{
			"valid_amount_00030,00EUR",
			getEmail(mailTemplate, "tests/amount/valid_amount_00030,00EUR.html", true),
			&data.Transaction{Base: 30, Fraction: 00},
			nil,
		},
		{
			"valid_amount_1.234,56EUR",
			getEmail(mailTemplate, "tests/amount/valid_amount_1.234,56EUR.html", true),
			&data.Transaction{Base: 1234, Fraction: 56},
			nil,
		},
		{
			"valid_amount_1234,56EUR",
			getEmail(mailTemplate, "tests/amount/valid_amount_1234,56EUR.html", true),
			&data.Transaction{Base: 1234, Fraction: 56},
			nil,
		},
		{
			"valid_amount_20EUR",
			getEmail(mailTemplate, "tests/amount/valid_amount_20EUR.html", true),
			&data.Transaction{Base: 20, Fraction: 00},
			nil,
		},
		{
			"valid_amount_whitespace",
			getEmail(mailTemplate, "tests/amount/valid_amount_whitespace.html", true),
			&data.Transaction{Base: 12, Fraction: 34},
			nil,
		},
		{
			"valid_amount_2.99USD",
			getEmail(mailTemplate, "tests/amount/valid_amount_2.99USD.html", true),
			&data.Transaction{Base: 2, Fraction: 99},
			nil,
		},
		{
			"valid_amount_12,345.67USD",
			getEmail(mailTemplate, "tests/amount/valid_amount_12,345.67USD.html", true),
			&data.Transaction{Base: 12345, Fraction: 67},
			nil,
		},
		{
			"valid_amount_1.00002USD",
			getEmail(mailTemplate, "tests/amount/valid_amount_1.00002USD.html", true),
			&data.Transaction{Base: 1, Fraction: 0},
			nil,
		},
	}

	for _, test := range testTable {
		parser := NewTransactionMailParser(nameAmountRegex)
		output, err := parser.GetTransactionInfo(test.inputMail)
		if !compareErrors(err, test.expectError) {
			t.Fatalf("GetTransactionInfo(%s) returned error %v, but should return with error %v", test.name, err, test.expectError)
		}
		if test.expectedOut == nil {
			return
		}
		if output.Base != test.expectedOut.Base || output.Fraction != test.expectedOut.Fraction {
			t.Fatalf("GetTransactionInfo(%s) returned amount (%v,%v), but should return (%v,%v)", test.name, output.Base, output.Fraction, test.expectedOut.Base, test.expectedOut.Fraction)
		}
	}
}

func TestGetTransactionInfoNotes(t *testing.T) {
	testTable := []TransactionTest{
		{
			"valid_note_my_note",
			getEmail(mailTemplate, "tests/note/valid_note_my_note.html", true),
			&data.Transaction{Note: "My Note"},
			nil,
		},
		{
			"valid_note_special_chars",
			getEmail(mailTemplate, "tests/note/valid_note_special_chars.html", true),
			&data.Transaction{Note: "我的笔记"},
			nil,
		},
	}
	for _, test := range testTable {
		parser := NewTransactionMailParser(nameAmountRegex)
		output, err := parser.GetTransactionInfo(test.inputMail)
		if !compareErrors(err, test.expectError) {
			t.Fatalf("GetTransactionInfo(%s) returned error %v, but should return with error %v", test.name, err, test.expectError)
		}
		if test.expectedOut == nil {
			return
		}
		if output.Note != test.expectedOut.Note {
			t.Fatalf("GetTransactionInfo(%s) returned note %v, but should return %v", test.name, output.Note, test.expectedOut.Note)
		}
	}
}

func TestGetTransactionInfoInvalid(t *testing.T) {
	testTable := []TransactionTest{
		{
			"no_name",
			getEmail(mailTemplate, "tests/invalid/no_name.html", true),
			nil,
			errors.New("Error while getting parser info no text in html matched parser pattern"),
		},
		{
			"no_amount",
			getEmail(mailTemplate, "tests/invalid/no_amount.html", true),
			nil,
			errors.New("Error while getting parser info no text in html matched parser pattern"),
		},
		{
			"not_base64_encoded",
			getEmail(mailTemplate, "tests/name/valid_name_two_words.html", false),
			nil,
			errors.New("Error while decoding base64 html illegal base64 data at input byte 0"),
		},
		{
			"no_texts_in_html",
			getEmail(mailTemplate, "tests/invalid/no_texts.html", true),
			nil,
			errors.New("Error while getting parser info no span text found"),
		},
		{
			"invalid_html",
			getEmail(mailTemplate, "tests/invalid/invalid_html.html", true),
			nil,
			errors.New("Error while getting parser info no span text found"),
		},
		{
			"no_note",
			getEmail(mailTemplate, "tests/invalid/no_note.html", true),
			nil,
			errors.New("Error while getting Note could not get image tag surrounding Note"),
		},
		{
			"no_currency",
			getEmail(mailTemplate, "tests/invalid/no_currency.html", true),
			nil,
			errors.New("Error while getting parser info no text in html matched parser pattern"),
		},
		{
			"invalid_currency",
			getEmail(mailTemplate, "tests/invalid/invalid_currency.html", true),
			nil,
			errors.New("Error while getting parser info no text in html matched parser pattern"),
		},
	}
	for _, test := range testTable {
		parser := NewTransactionMailParser(nameAmountRegex)
		output, err := parser.GetTransactionInfo(test.inputMail)
		if !compareErrors(err, test.expectError) || output != nil {
			t.Fatalf("GetTransactionInfo(%s) returned %v, %v but should return %v, %v", test.name, output, err, nil, test.expectError)
		}
	}
}

func compareErrors(err1, err2 error) bool {
	if err1 != nil && err2 != nil {
		return err1.Error() == err2.Error()
	}
	return err1 == err2
}

func getEmail(mailTemplate *template.Template, mailBodyFileName string, base64Encode bool) parsemail.Email {
	mailBodyText, err := ioutil.ReadFile(mailBodyFileName)
	mailBodyString := string(mailBodyText)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if base64Encode {
		mailBodyString = b64.StdEncoding.EncodeToString(mailBodyText)
	}
	err = mailTemplate.Execute(buf, mailBodyString)
	if err != nil {
		panic(err)
	}
	email, err := parsemail.Parse(buf)
	if err != nil {
		panic(err)
	}
	return email
}
