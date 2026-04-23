package main

import "testing"

func TestIndexToCell(t *testing.T) {
	tests := []struct {
		name        string
		rowIndex    int
		colIndex    int
		want    string
		wantErr bool
	}{
		{"A1", 0, 0, "A1", false},
		{"B1", 0, 1, "B1", false},
		{"A2", 1, 0, "A2", false},
		{"B2", 1, 1, "B2", false},

		{"XFD1", 0, 16_383, "XFD1", false},
		{"A1048576", 1_048_575, 0, "A1048576", false},
		{"XFD1048576", 1_048_575, 16_383, "XFD1048576", false},

		{"Negative Error1", 0, -1, "", true},
		{"Negative Error2", -1, 0, "", true},
		{"Negative Error3", -1, -1, "", true},

		{"Too Large Error1", 0, 16_384, "", true},
		{"Too Large Error2", 1_048_576, 0, "", true},
		{"Too Large Error3", 1_048_576, 16_384, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := indexToCell(tt.rowIndex, tt.colIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("indexToCell(%d, %d) error = %v, want error = %t", tt.rowIndex, tt.colIndex, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("indexToCell(%d, %d) = %s, want %s", tt.rowIndex, tt.colIndex, got, tt.want)
			}
		})
	}
}

func TestRoomStringToNumbers(t *testing.T){
	tests := []struct {
		name       string
		roomNumber string
		wantSlice  []int
		wantErr    bool
	}{
		{"Valid Number1", "101", []int{101}, false},
		{"Valid Number2", "149", []int{149}, false},
		{"Valid Number2", "949", []int{949}, false},

		{"Valid Range1", "101~102", []int{101, 102}, false},
		{"Valid Range2", "101~103", []int{101, 102, 103}, false},
		{"Valid Range3", "101~149", []int{101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149}, false},

		{"Valid MultiNumbers1", "101,103", []int{101, 103}, false},

		{"Valid MultiRange1", "101,103~105", []int{101, 103, 104, 105}, false},
		{"Valid MultiRange2", "101~103,105", []int{101, 102, 103, 105}, false},
		{"Valid MultiRange3", "101~103,105~107", []int{101, 102, 103, 105, 106, 107}, false},
		{"Valid MultiRange4", "101,103~105,107~109", []int{101, 103, 104, 105, 107, 108, 109}, false},
		{"Valid MultiRange5", "101~103,105,107~109", []int{101, 102, 103, 105, 107, 108, 109}, false},
		{"Valid MultiRange6", "101~103,105~107,109", []int{101, 102, 103, 105, 106, 107, 109}, false},
		{"Valid MultiRange7", "101~103,105~107,109~111", []int{101, 102, 103, 105, 106, 107, 109, 110, 111}, false},

		{"Valid Number Space1", " 101", []int{101}, false},
		{"Valid Number Space2", "101 ", []int{101}, false},
		{"Valid Number Space3", " 101 ", []int{101}, false},

		{"Valid Numbers Space1", " 101,103", []int{101, 103}, false},
		{"Valid Numbers Space2", "101 ,103", []int{101, 103}, false},
		{"Valid Numbers Space3", "101, 103", []int{101, 103}, false},
		{"Valid Numbers Space4", "101,103 ", []int{101, 103}, false},
		{"Valid Numbers Space5", " 101 ,103", []int{101, 103}, false},
		{"Valid Numbers Space6", " 101, 103", []int{101, 103}, false},
		{"Valid Numbers Space7", " 101,103 ", []int{101, 103}, false},
		{"Valid Numbers Space8", "101 , 103", []int{101, 103}, false},
		{"Valid Numbers Space9", "101 ,103 ", []int{101, 103}, false},
		{"Valid Numbers Space10", "101, 103 ", []int{101, 103}, false},
		{"Valid Numbers Space11", " 101 , 103", []int{101, 103}, false},
		{"Valid Numbers Space12", " 101 ,103 ", []int{101, 103}, false},
		{"Valid Numbers Space13", " 101, 103 ", []int{101, 103}, false},
		{"Valid Numbers Space14", "101 , 103 ", []int{101, 103}, false},
		{"Valid Numbers Space15", " 101 , 103 ", []int{101, 103}, false},

		{"Valid Range Space1", " 101~103", []int{101, 102, 103}, false},
		{"Valid Range Space2", "101~103 ", []int{101, 102, 103}, false},
		{"Valid Range Space3", " 101~103 ", []int{101, 102, 103}, false},
		{"Valid Range Space4", "101 ~103", []int{101, 102, 103}, false},
		{"Valid Range Space5", "101~ 103", []int{101, 102, 103}, false},
		{"Valid Range Space6", "101 ~ 103", []int{101, 102, 103}, false},
		{"Valid Range Space7", " 101 ~ 103 ", []int{101, 102, 103}, false},

		{"Invalid Number1", "01", []int{}, true},
		{"Invalid Number2", "1001", []int{}, true},
		{"Invalid Number3", "001", []int{}, true},
		{"Invalid Number4", "049", []int{}, true},
		{"Invalid Number5", "100", []int{}, true},
		{"Invalid Number6", "150", []int{}, true},
		{"Invalid Number7", "900", []int{}, true},
		{"Invalid Number8", "950", []int{}, true},

		{"Invalid Range1", "~101", []int{}, true},
		{"Invalid Range2", "101~", []int{}, true},
		{"Invalid Range3", "001~003", []int{}, true},
		{"Invalid Range4", "001~101", []int{}, true},
		{"Invalid Range5", "100~102", []int{}, true},
		{"Invalid Range6", "150~148", []int{}, true},
		{"Invalid Range7", "101~001", []int{}, true},
		{"Invalid Range8", "102~100", []int{}, true},
		{"Invalid Range9", "148~150", []int{}, true},
		{"Invalid Range10", "103~101", []int{}, true},

		{"Invalid Character1", "10a", []int{}, true},
		{"Invalid Character2", "101~20a", []int{}, true},
		{"Invalid Character3", "abc", []int{}, true},
		{"Invalid Character4", "101,abc,103", []int{}, true},
		{"Invalid Character5", "101~abc", []int{}, true},
		{"Invalid Character6", "abc~103", []int{}, true},
		{"Invalid Character7", "101-103", []int{}, true},
		{"Invalid Character8", "101, 102~103, xyz", []int{}, true},
		{"Invalid Character9", "10 1", []int{}, true},
		{"Invalid Character10", "１０１", []int{}, true},
		{"Invalid Character11", "101, , 103", []int{}, true},
		{"Invalid Character12", "101,,103", []int{}, true},
		{"Invalid Character13", "101~103~105", []int{}, true},
		{"Invalid Character14", "101.5", []int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlice, err := roomStringToNumbers(tt.roomNumber)
			if (err != nil) != tt.wantErr {
				t.Fatalf("roomStringToNumbers(%s) error = %v, want error = %t", tt.roomNumber, err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(gotSlice) != len(tt.wantSlice) {
					t.Fatalf("roomStringToNumbers(%s) = %v, want %v", tt.roomNumber, gotSlice, tt.wantSlice)
				}
				for i := range gotSlice {
					if gotSlice[i] != tt.wantSlice[i] {
						t.Errorf("roomStringToNumbers(%s) = %v, want %v", tt.roomNumber, gotSlice, tt.wantSlice)
						break
					}
				}
			}
		})
	}
}
func TestIsTask(t *testing.T){
	tests := []struct {
		name        string
		task        string
		expected    bool
	}{
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTask(tt.task); got != tt.expected {
				t.Errorf("isTask(%s) = %t, want %t", tt.task, got, tt.expected)
			}
		})
	}
}
func TestCheckExcelSyntax(t *testing.T){
	tests := []struct {
		name        string
		excel       [][]string
		expected    error
		expectedErr bool
	}{
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
		})
	}
}