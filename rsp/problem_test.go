package rsp

import (
	"encoding/json"
	"testing"

	"go-slim.dev/v"
)

func TestProblemCreation(t *testing.T) {
	tests := []struct {
		name       string
		problem    *Problem
		wantLabel  string
		wantCode   string
		wantMsg    string
		wantNested bool
	}{
		{
			name: "Simple problem",
			problem: &Problem{
				Label:   "email",
				Code:    "INVALID_FORMAT",
				Message: "Invalid email format",
			},
			wantLabel:  "email",
			wantCode:   "INVALID_FORMAT",
			wantMsg:    "Invalid email format",
			wantNested: false,
		},
		{
			name: "Problem with nested problems",
			problem: &Problem{
				Label:   "user",
				Code:    "VALIDATION_FAILED",
				Message: "User validation failed",
				Problems: Problems{
					"email": {
						{
							Label:   "email",
							Code:    "INVALID_FORMAT",
							Message: "Invalid email format",
						},
					},
				},
			},
			wantLabel:  "user",
			wantCode:   "VALIDATION_FAILED",
			wantMsg:    "User validation failed",
			wantNested: true,
		},
		{
			name: "Empty problem",
			problem: &Problem{
				Label:   "",
				Code:    "",
				Message: "",
			},
			wantLabel:  "",
			wantCode:   "",
			wantMsg:    "",
			wantNested: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.problem == nil {
				t.Error("Problem creation: expected not nil, got nil")
				return
			}

			// Verify the Problem struct was created with correct values
			if tt.problem.Label != tt.wantLabel {
				t.Errorf("Problem label = %q, want %q", tt.problem.Label, tt.wantLabel)
			}

			if tt.problem.Code != tt.wantCode {
				t.Errorf("Problem code = %q, want %q", tt.problem.Code, tt.wantCode)
			}

			if tt.problem.Message != tt.wantMsg {
				t.Errorf("Problem message = %q, want %q", tt.problem.Message, tt.wantMsg)
			}

			// Check if nested problems exist as expected
			hasNested := len(tt.problem.Problems) > 0
			if hasNested != tt.wantNested {
				t.Errorf("Problem has nested problems = %v, want %v", hasNested, tt.wantNested)
			}

			// If nested problems are expected, verify their structure
			if tt.wantNested && hasNested {
				if _, exists := tt.problem.Problems["email"]; !exists {
					t.Error("Expected nested email problem not found")
				}
			}
		})
	}
}

func TestProblemsAdd(t *testing.T) {
	problems := make(Problems)

	// Add first problem to a field
	problem1 := &Problem{
		Label:   "email",
		Code:    "INVALID_FORMAT",
		Message: "Invalid email format",
	}

	problems.Add(problem1)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	emailProblems, exists := problems["email"]
	if !exists {
		t.Error("Email field should exist in problems")
		return
	}

	if len(emailProblems) != 1 {
		t.Errorf("Email problems length = %v, want 1", len(emailProblems))
	}

	if emailProblems[0] != problem1 {
		t.Error("Added problem not found in email problems")
	}

	// Add second problem to the same field
	problem2 := &Problem{
		Label:   "email",
		Code:    "BLACKLISTED",
		Message: "Email domain is not allowed",
	}

	problems.Add(problem2)

	if len(problems) != 1 {
		t.Errorf("Problems length should remain 1, got %v", len(problems))
	}

	emailProblems = problems["email"]
	if len(emailProblems) != 2 {
		t.Errorf("Email problems length = %v, want 2", len(emailProblems))
	}

	if emailProblems[1] != problem2 {
		t.Error("Second problem not found in email problems")
	}

	// Add problem to a different field
	problem3 := &Problem{
		Label:   "password",
		Code:    "TOO_SHORT",
		Message: "Password must be at least 8 characters",
	}

	problems.Add(problem3)

	if len(problems) != 2 {
		t.Errorf("Problems length = %v, want 2", len(problems))
	}

	passwordProblems, exists := problems["password"]
	if !exists {
		t.Error("Password field should exist in problems")
		return
	}

	if len(passwordProblems) != 1 {
		t.Errorf("Password problems length = %v, want 1", len(passwordProblems))
	}
}

func TestProblemsAddError(t *testing.T) {
	problems := make(Problems)

	// Create a validation error with field context
	err := v.Value("invalid-email", "email", "Email").
		Custom("INVALID_FORMAT", func(val any) any {
			return false // Trigger validation error
		}, v.ErrorFormat("Invalid email format")).
		Validate()

	validationErr, _ := err.(*v.Error)
	problems.AddError(validationErr)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	emailProblems, exists := problems["email"]
	if !exists {
		t.Error("Email field should exist in problems")
		return
	}

	if len(emailProblems) != 1 {
		t.Errorf("Email problems length = %v, want 1", len(emailProblems))
	}

	addedProblem := emailProblems[0]
	if addedProblem.Label != "email" {
		t.Errorf("Problem label = %v, want email", addedProblem.Label)
	}

	if addedProblem.Code != "INVALID_FORMAT" {
		t.Errorf("Problem code = %v, want INVALID_FORMAT", addedProblem.Code)
	}

	if addedProblem.Message != "Invalid email format" {
		t.Errorf("Problem message = %v, want 'Invalid email format'", addedProblem.Message)
	}

	if addedProblem.Problems != nil {
		t.Error("Problem should not have nested problems")
	}
}

func TestProblemsMultipleAddError(t *testing.T) {
	problems := make(Problems)

	// Add multiple validation errors with field context
	errors := []*v.Error{
		v.Value("invalid", "email", "Email").
			Custom("INVALID_FORMAT", func(val any) any {
				return false
			}, v.ErrorFormat("Invalid email format")).Validate().(*v.Error),
		v.Value("bad@domain.com", "email", "Email").
			Custom("BLACKLISTED", func(val any) any {
				return false
			}, v.ErrorFormat("Email domain is not allowed")).Validate().(*v.Error),
		v.Value("123", "password", "Password").
			Custom("TOO_SHORT", func(val any) any {
				return false
			}, v.ErrorFormat("Password must be at least 8 characters")).Validate().(*v.Error),
		v.Value(15, "age", "Age").
			Custom("UNDERAGE", func(val any) any {
				return false
			}, v.ErrorFormat("Must be at least 18 years old")).Validate().(*v.Error),
	}

	for _, err := range errors {
		problems.AddError(err)
	}

	if len(problems) != 3 {
		t.Errorf("Problems length = %v, want 3", len(problems))
	}

	// Check email field has 2 problems
	emailProblems := problems["email"]
	if len(emailProblems) != 2 {
		t.Errorf("Email problems length = %v, want 2", len(emailProblems))
	}

	// Check password field has 1 problem
	passwordProblems := problems["password"]
	if len(passwordProblems) != 1 {
		t.Errorf("Password problems length = %v, want 1", len(passwordProblems))
	}

	// Check age field has 1 problem
	ageProblems := problems["age"]
	if len(ageProblems) != 1 {
		t.Errorf("Age problems length = %v, want 1", len(ageProblems))
	}
}

func TestProblemJSONSerialization(t *testing.T) {
	problems := Problems{
		"email": {
			{
				Label:   "email",
				Code:    "INVALID_FORMAT",
				Message: "Invalid email format",
			},
		},
		"password": {
			{
				Label:   "password",
				Code:    "TOO_SHORT",
				Message: "Password must be at least 8 characters",
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(problems)
	if err != nil {
		t.Errorf("JSON marshaling error = %v", err)
		return
	}

	// Deserialize back
	var unmarshaled Problems
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("JSON unmarshaling error = %v", err)
		return
	}

	// Check if structure is preserved
	if len(unmarshaled) != len(problems) {
		t.Errorf("Unmarshaled problems length = %v, want %v", len(unmarshaled), len(problems))
	}

	// Check email problems
	emailProblems := unmarshaled["email"]
	if len(emailProblems) != 1 {
		t.Errorf("Email problems length = %v, want 1", len(emailProblems))
	}

	if emailProblems[0].Code != "INVALID_FORMAT" {
		t.Errorf("Email problem code = %v, want INVALID_FORMAT", emailProblems[0].Code)
	}

	// Check password problems
	passwordProblems := unmarshaled["password"]
	if len(passwordProblems) != 1 {
		t.Errorf("Password problems length = %v, want 1", len(passwordProblems))
	}

	if passwordProblems[0].Code != "TOO_SHORT" {
		t.Errorf("Password problem code = %v, want TOO_SHORT", passwordProblems[0].Code)
	}
}

func TestSingleProblemJSONSerialization(t *testing.T) {
	problem := Problem{
		Label:   "email",
		Code:    "INVALID_FORMAT",
		Message: "Invalid email format",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(problem)
	if err != nil {
		t.Errorf("JSON marshaling error = %v", err)
		return
	}

	// Deserialize back
	var unmarshaled Problem
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("JSON unmarshaling error = %v", err)
		return
	}

	// Check if fields are preserved (except Label which has json:"-")
	if unmarshaled.Code != problem.Code {
		t.Errorf("Problem code = %v, want %v", unmarshaled.Code, problem.Code)
	}

	if unmarshaled.Message != problem.Message {
		t.Errorf("Problem message = %v, want %v", unmarshaled.Message, problem.Message)
	}

	// Label should not be serialized due to json:"-" tag
	if unmarshaled.Label != "" {
		t.Errorf("Problem label should not be serialized, got %v", unmarshaled.Label)
	}
}

func TestProblemsWithComplexStructure(t *testing.T) {
	problems := make(Problems)

	// Add a problem with nested problems
	nestedProblems := make(Problems)
	nestedProblems.Add(&Problem{
		Label:   "password",
		Code:    "TOO_SHORT",
		Message: "Password too short",
	})
	nestedProblems.Add(&Problem{
		Label:   "password",
		Code:    "NO_UPPERCASE",
		Message: "Password must contain uppercase letters",
	})

	complexProblem := &Problem{
		Label:    "some_validation",
		Code:     "some",
		Message:  "At least one of these conditions must be satisfied",
		Problems: nestedProblems,
	}

	problems.Add(complexProblem)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	someProblems := problems["some_validation"]
	if len(someProblems) != 1 {
		t.Errorf("Some validation problems length = %v, want 1", len(someProblems))
	}

	mainProblem := someProblems[0]
	if mainProblem.Code != "some" {
		t.Errorf("Main problem code = %v, want some", mainProblem.Code)
	}

	if len(mainProblem.Problems) != 1 {
		t.Errorf("Nested problems length = %v, want 1", len(mainProblem.Problems))
	}

	passwordProblems := mainProblem.Problems["password"]
	if len(passwordProblems) != 2 {
		t.Errorf("Nested password problems length = %v, want 2", len(passwordProblems))
	}
}

func TestCollectProblemSimpleError(t *testing.T) {
	problems := make(Problems)

	// Create a simple error with field context
	err := v.Value("invalid", "email", "Email").
		Custom("INVALID_FORMAT", func(val any) any {
			return false
		}, v.ErrorFormat("Invalid email format")).
		Validate()

	simpleErr, _ := err.(*v.Error)
	collectProblem(problems, simpleErr)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	emailProblems := problems["email"]
	if len(emailProblems) != 1 {
		t.Errorf("Email problems length = %v, want 1", len(emailProblems))
	}

	if emailProblems[0].Code != "INVALID_FORMAT" {
		t.Errorf("Problem code = %v, want INVALID_FORMAT", emailProblems[0].Code)
	}
}

func TestCollectProblemWithInternalError(t *testing.T) {
	problems := make(Problems)

	// Create a simple error with field context
	err := v.Value("invalid", "field", "Field").
		Custom("INVALID", func(val any) any {
			return false
		}, v.ErrorFormat("Field validation failed")).
		Validate()

	mainErr, _ := err.(*v.Error)
	collectProblem(problems, mainErr)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	fieldProblems := problems["field"]
	if len(fieldProblems) != 1 {
		t.Errorf("Field problems length = %v, want 1", len(fieldProblems))
	}

	if fieldProblems[0].Code != "INVALID" {
		t.Errorf("Problem code = %v, want INVALID", fieldProblems[0].Code)
	}
}

func TestProblemsWithEmptyField(t *testing.T) {
	problems := make(Problems)

	// Add problem with empty field name
	problem := &Problem{
		Label:   "",
		Code:    "GENERAL_ERROR",
		Message: "A general error occurred",
	}

	problems.Add(problem)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	emptyFieldProblems := problems[""]
	if len(emptyFieldProblems) != 1 {
		t.Errorf("Empty field problems length = %v, want 1", len(emptyFieldProblems))
	}

	if emptyFieldProblems[0].Code != "GENERAL_ERROR" {
		t.Errorf("Problem code = %v, want GENERAL_ERROR", emptyFieldProblems[0].Code)
	}
}

func TestProblemsWithSpecialCharacters(t *testing.T) {
	problems := make(Problems)

	// Add problem with special characters in label and message
	problem := &Problem{
		Label:   "field-with-special.chars",
		Code:    "SPECIAL_CHARS_ERROR",
		Message: "Error with special chars: !@#$%^&*()",
	}

	problems.Add(problem)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	fieldProblems := problems["field-with-special.chars"]
	if len(fieldProblems) != 1 {
		t.Errorf("Field problems length = %v, want 1", len(fieldProblems))
	}

	expectedMsg := "Error with special chars: !@#$%^&*()"
	if fieldProblems[0].Message != expectedMsg {
		t.Errorf("Problem message = %s, want %s", fieldProblems[0].Message, expectedMsg)
	}

	// Test JSON serialization with special characters
	jsonData, err := json.Marshal(problems)
	if err != nil {
		t.Errorf("JSON marshaling with special chars error = %v", err)
		return
	}

	var unmarshaled Problems
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("JSON unmarshaling with special chars error = %v", err)
		return
	}

	if len(unmarshaled) != 1 {
		t.Errorf("Unmarshaled problems with special chars length = %v, want 1", len(unmarshaled))
	}
}

func TestProblemsUnicodeSupport(t *testing.T) {
	problems := make(Problems)

	// Add problem with Unicode characters
	problem := &Problem{
		Label:   "用户名",
		Code:    "INVALID_USERNAME",
		Message: "用户名格式无效",
	}

	problems.Add(problem)

	if len(problems) != 1 {
		t.Errorf("Problems length = %v, want 1", len(problems))
	}

	usernameProblems := problems["用户名"]
	if len(usernameProblems) != 1 {
		t.Errorf("Username problems length = %v, want 1", len(usernameProblems))
	}

	if usernameProblems[0].Message != "用户名格式无效" {
		t.Errorf("Problem message = %v, want '用户名格式无效'", usernameProblems[0].Message)
	}

	// Test JSON serialization with Unicode
	jsonData, err := json.Marshal(problems)
	if err != nil {
		t.Errorf("JSON marshaling with Unicode error = %v", err)
		return
	}

	var unmarshaled Problems
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("JSON unmarshaling with Unicode error = %v", err)
		return
	}

	if len(unmarshaled) != 1 {
		t.Errorf("Unmarshaled Unicode problems length = %v, want 1", len(unmarshaled))
	}
}

func TestAddErrorWithNilProblem(t *testing.T) {
	problems := make(Problems)

	// This test ensures that the function handles nil input gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddError should not panic with nil input, got panic: %v", r)
		}
	}()

	// This should not panic (though in practice, you shouldn't pass nil)
	var err *v.Error = nil
	problems.AddError(err)

	// Problems should remain unchanged
	if len(problems) != 0 {
		t.Errorf("Problems length = %v, want 0", len(problems))
	}
}

// Benchmarks for problem handling
func BenchmarkProblemAdd(b *testing.B) {
	problems := make(Problems)
	problem := &Problem{
		Label:   "email",
		Code:    "INVALID_FORMAT",
		Message: "Invalid email format",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		problems.Add(problem)
	}
}

func BenchmarkProblemMultipleAdd(b *testing.B) {
	problem1 := &Problem{
		Label:   "email",
		Code:    "INVALID_FORMAT",
		Message: "Invalid email format",
	}
	problem2 := &Problem{
		Label:   "password",
		Code:    "TOO_SHORT",
		Message: "Password must be at least 8 characters",
	}
	problem3 := &Problem{
		Label:   "age",
		Code:    "UNDERAGE",
		Message: "Must be at least 18 years old",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		problems := make(Problems)
		problems.Add(problem1)
		problems.Add(problem2)
		problems.Add(problem3)
	}
}

func BenchmarkProblemsJSONMarshal(b *testing.B) {
	problems := make(Problems)
	problems.Add(&Problem{
		Label:   "email",
		Code:    "INVALID_FORMAT",
		Message: "Invalid email format",
	})
	problems.Add(&Problem{
		Label:   "password",
		Code:    "TOO_SHORT",
		Message: "Password must be at least 8 characters",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(problems)
	}
}

func BenchmarkParallelProblems(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			problems := make(Problems)
			problems.Add(&Problem{
				Label:   "email",
				Code:    "INVALID_FORMAT",
				Message: "Invalid email format",
			})
		}
	})
}

func BenchmarkProblemsCreationAndAdd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		problems := make(Problems)
		problems.Add(&Problem{
			Label:   "field1",
			Code:    "REQUIRED",
			Message: "This field is required",
		})
		problems.Add(&Problem{
			Label:   "field2",
			Code:    "INVALID_FORMAT",
			Message: "Invalid format",
		})
	}
}

func BenchmarkProblemsWithErrors(b *testing.B) {
	problem1 := &Problem{
		Label:   "field1",
		Code:    "REQUIRED",
		Message: "Field is required",
	}
	problem2 := &Problem{
		Label:   "field2",
		Code:    "INVALID_FORMAT",
		Message: "Invalid format",
	}
	problem3 := &Problem{
		Label:   "field3",
		Code:    "TOO_SHORT",
		Message: "Too short",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newProblems := make(Problems)
		newProblems.Add(problem1)
		newProblems.Add(problem2)
		newProblems.Add(problem3)
	}
}
