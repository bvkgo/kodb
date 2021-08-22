package internal

import "testing"

func TestDataTypeMap(t *testing.T) {
	type Employee struct {
		Email string `kodb:"index"`
	}
	if err := Register("Employee", new(Employee)); err != nil {
		t.Fatal(err)
	}
	if err := Register("Employee", Employee{}); err == nil {
		t.Fatalf("duplicate type names must not be allowed")
	}
	if err := Register("EmployeeV2", Employee{}); err == nil {
		t.Fatal("same type cannot be registered multiple times")
	}
	if err := Register("EmployeeV2", new(Employee)); err == nil {
		t.Fatal("same element type cannot be registered multiple times")
	}
	if _, err := GetDataType(Employee{}); err != nil {
		t.Fatal(err)
	}
	if _, err := GetDataType(new(Employee)); err != nil {
		t.Fatal(err)
	}
}
