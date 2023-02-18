package types

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

const (
	test     = String("12345")
	testI8   = String("123")
	testF    = String("123.4556")
	testF2   = String("12345.56")
	testBool = String("true")
)

type TestStruct struct {
	Name string
	Age  int
}

func TestString(t *testing.T) {
	if test != "12345" {
		t.Fatal("Error evaluating:", test)
	}
}

func TestStringBytes(t *testing.T) {
	if !bytes.Equal(test.Bytes(), []byte("12345")) {
		t.Fatal("Error evaluating:", test, "Bytes()")
	}
}

func TestStringInt(t *testing.T) {
	if test.Int() != 12345 {
		t.Fatal("Error evaluating:", test, "Int()")
	}
}

func TestStringInt8(t *testing.T) {
	if testI8.Int().Int8() != 123 {
		t.Fatal("Error evaluating:", testI8, "Int8()")
	}
}

func TestStringInt16(t *testing.T) {
	if test.Int().Int16() != 12345 {
		t.Fatal("Error evaluating:", test, "Int16()")
	}
}

func TestStringFloat(t *testing.T) {
	if testF.Float() != 123.4556 {
		t.Fatal("Error evaluating:", testF, "Float()")
	}
}

func TestStringUpdate(t *testing.T) {
	if test.Update(func(s string) string { return strings.Replace(s, "5", "567890", 1) }) != "1234567890" {
		t.Fatal("Error evaluating:", test, "Update()")
	}
}

func TestFloatToFixed(t *testing.T) {
	f64 := Float(10)
	expected := String("10.00")

	if f64.FixedStr(2) != expected {
		t.Fatal("Error evaluating:", f64, "FixedStr()", expected)
	}

	f64 = String("10.99884").Float()
	expected = "10.998"

	if f64.FixedStr(3) != expected {
		t.Fatal("Error evaluating:", f64, "FixedStr()", expected)
	}

	expected = "10.99884000"

	if f64.FixedStr(8) != expected {
		t.Fatal("Error evaluating:", f64, "FixedStr()", expected)
	}

	f32 := Float(f64)
	expected = "10.998840"

	if f32.FixedStr(6) != expected {
		t.Fatal("Error evaluating:", f32, "FixedStr()", expected)
	}

}

func TestStringBool(t *testing.T) {
	if testBool.Bool() != true {
		t.Fatal("Error evaluating:", testBool, "Bool()")
	}
}

func TestStringTrueInt(t *testing.T) {
	if testBool.Int() != 0 {
		t.Fatal("Error evaluating:", testBool, "Int()")
	}
}

func TestStringIntFloatIntStringBytesString(t *testing.T) {
	result := testF2.Int().Bytes().String().Int().String().Float()

	if result != 12345 {
		t.Fatal("Error evaluating:", test, "Int().Bytes().String().Int().String().Float()", "result:", result, "expected", 12345)
	}
}

func TestStringMask(t *testing.T) {
	memberCode := String("testjus123")
	maskedCode := string(memberCode.Mask("*", 4, 4))
	expected := "****s123"

	if maskedCode != expected {
		t.Fatal("Error evaluating: ", memberCode, " Mask(): ", maskedCode, " expected: ", expected)
	}
}

func TestArrayMap(t *testing.T) {
	array := Array[any]{"string", 1, 123}
	containValue := 1

	if !array.Constains(containValue) {
		t.Fatal("Error evaluating:", array, "Constains()", containValue)
	}
}

func TestSHA256(t *testing.T) {
	intBytes := Int(14).Bytes()
	intExpected := "8527a891e224136950ff32ca212b45bc93f69fbb801c3b1ebedac52775f99e61"

	if intBytes.SHA256() != intExpected {
		t.Fatal("Error evaluating:", 14, "SHA256()", intExpected)
	}

	strBytes := String(`"14"`).Bytes()
	strEXpected := "789d9e7dc67f986660d9b28f5946cce995f9b145e4c2ef61828679cee07bb316"

	if strBytes.SHA256() != strEXpected {
		t.Fatal("Error evaluating:", 14, "SHA256()", strEXpected)
	}
}

func TestMD5(t *testing.T) {
	intBytes := String("kamote").Bytes()
	intExpected := "9d8f06fa59c5050be3d7462a783689a7"

	if intBytes.MD5() != intExpected {
		t.Fatal("Error evaluating:", 14, "MD5()", intExpected)
	}
}

func TestArray(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	aa := Array[int](a)
	aam := aa.Map(func(value int) any {
		return fmt.Sprint(value)
	})
	aExpected := "2"

	if aam[1] != aExpected {
		t.Fatal("Error evaluating:", aa, "Map() - index 1 - string", aExpected)
	}
	aamr := Array[String]{}
	aamrExpected := Array[String]{"1", "2", "3", "4", "5"}
	aam.Foreach(func(_ int, value any) {
		aamr = append(aamr, String(value.(string)))
	})

	if !aamr.IsEqualTo(aamrExpected) {
		t.Fatal("Error evaluating:", aamr, "IsEqualTo()", aamrExpected)
	}

	t1 := Array[any]{"1", "2", "3", "4", TestStruct{Name: "Justine", Age: 31}}
	t2 := Array[any]{"1", "2", "3", "4", TestStruct{Name: "Justine", Age: 31}}

	if !t1.IsEqualTo(t2) {
		t.Fatal("Error evaluating:", t1, "IsEqualTo()", t2)
	}

	t3 := Array[any]{String("1"), String("2")}
	t4 := Array[String]{"1", "2"}

	if !t3.IsEqualTo(t4.Map(func(value String) any { return value })) {
		t.Fatal("Error evaluating:", t3, "IsEqualTo()", t4)
	}
}

func TestArrayPop(t *testing.T) {
	arr := Array[string]{"potato", "tomato", "cabbage", "petchay"}
	index := 1
	fmt.Println("TestArrayPop index: ", index)
	fmt.Println("TestArrayPop before: ", arr.ToRaw())
	fmt.Println("TestArrayPop value: ", arr.PopIndex(index))
	fmt.Println("TestArrayPop after: ", arr.ToRaw())
}

func TestOdds(t *testing.T) {
	euroToUK := Odds(34.78869).EuroToUK(4)
	expectedEuroToUK := Odds(33.7886)

	if euroToUK != expectedEuroToUK {
		t.Fatal("Error evaluating euroToUK: ", euroToUK, " expectedEuroToUK: ", expectedEuroToUK)
	}
	euroToUS := Odds(2.67).EuroToUS(2)
	expectedEuroToUS := Odds(167)

	if euroToUS != expectedEuroToUS {
		t.Fatal("Error evaluating euroToUS: ", euroToUS, " expectedEuroToUS: ", expectedEuroToUS)
	}
	euroToUS = Odds(34.9887).EuroToUS(2)
	expectedEuroToUS = Odds(3398.87)

	if euroToUS != expectedEuroToUS {
		t.Fatal("Error evaluating euroToUS: ", euroToUS, " expectedEuroToUS: ", expectedEuroToUS)
	}
	usToEuro := Odds(1985).USToEuro(2)
	expectedUSTOEuro := Odds(20.8500)

	if usToEuro != expectedUSTOEuro {
		t.Fatal("Error evaluating usToEuro: ", usToEuro, " expectedUSTOEuro: ", expectedUSTOEuro)
	}

	usToEuro = Odds(105).USToEuro(2)
	expectedUSTOEuro = Odds(2.04)

	if usToEuro != expectedUSTOEuro {
		t.Fatal("Error evaluating usToEuro: ", usToEuro, " expectedUSTOEuro: ", expectedUSTOEuro)
	}
	euroToMalay := Odds(1.667887).EuroToMalay(4)
	expectedEuroToMalay := Odds(0.6678)

	if euroToMalay != expectedEuroToMalay {
		t.Fatal("Error evaluating euroToMalay: ", euroToMalay, " expectedEuroToMalay: ", expectedEuroToMalay)
	}
	euroToMalay = Odds(134.667887).EuroToMalay(4)
	expectedEuroToMalay = Odds(-0.0074)

	if euroToMalay != expectedEuroToMalay {
		t.Fatal("Error evaluating euroToMalay: ", euroToMalay, " expectedEuroToMalay: ", expectedEuroToMalay)
	}
	malayToEuro := Odds(-0.0075).MalayToEuro(4)
	expectedMalayToEuro := Odds(134.3333)

	if malayToEuro != expectedMalayToEuro {
		t.Fatal("Error evaluating malayToEuro: ", malayToEuro, " expectedMalayToEuro: ", expectedMalayToEuro)
	}
	euroToHK := Odds(45.789979).EuroToHK(4)
	expectedEuroToHK := Odds(44.7899)

	if euroToHK != expectedEuroToHK {
		t.Fatal("Error evaluating euroToHK: ", euroToHK, " expectedEuroToHK: ", expectedEuroToHK)
	}
	euroToHK = Odds(1.564423).EuroToHK(4)
	expectedEuroToHK = Odds(0.5644)

	if euroToHK != expectedEuroToHK {
		t.Fatal("Error evaluating euroToHK: ", euroToHK, " expectedEuroToHK: ", expectedEuroToHK)
	}
	hkToEuro := Odds(0.01).HKToEuro(4)
	expectedHKToEuro := Odds(1.0100)

	if hkToEuro != expectedHKToEuro {
		t.Fatal("Error evaluating hkToEuro: ", hkToEuro, " expectedHKToEuro: ", expectedHKToEuro)
	}
	hkToEuro = Odds(9999).HKToEuro(4)
	expectedHKToEuro = Odds(10000.0000)

	if hkToEuro != expectedHKToEuro {
		t.Fatal("Error evaluating hkToEuro: ", hkToEuro, " expectedHKToEuro: ", expectedHKToEuro)
	}
	euroToIndo := Odds(45.668855).EuroToIndo(4)
	expectedEuroToIndo := Odds(44.6688)

	if euroToIndo != expectedEuroToIndo {
		t.Fatal("Error evaluating euroToIndo: ", euroToIndo, " expectedEuroToIndo: ", expectedEuroToIndo)
	}
	euroToIndo = Odds(1.06).EuroToIndo(4)
	expectedEuroToIndo = Odds(-16.6666)

	if euroToIndo != expectedEuroToIndo {
		t.Fatal("Error evaluating euroToIndo: ", euroToIndo, " expectedEuroToIndo: ", expectedEuroToIndo)
	}
	indoToEuro := Odds(-1.5).IndoToEuro(4)
	expectedIndoToEuro := Odds(1.6666)

	if indoToEuro != expectedIndoToEuro {
		t.Fatal("Error evaluating indoToEuro: ", indoToEuro, " expectedIndoToEuro: ", expectedIndoToEuro)
	}
	indoToEuro = Odds(1.584789).IndoToEuro(2)
	expectedIndoToEuro = Odds(2.5800)

	if indoToEuro != expectedIndoToEuro {
		t.Fatal("Error evaluating indoToEuro: ", indoToEuro, " expectedIndoToEuro: ", expectedIndoToEuro)
	}
	expectedIndoToEuroStr := "2.58000"

	if indoToEuro.String(5) != expectedIndoToEuroStr {
		t.Fatal("Error evaluating indoToEuro.String(5): ", indoToEuro.String(5), " expectedIndoToEuroStr: ", expectedIndoToEuroStr)
	}
}
