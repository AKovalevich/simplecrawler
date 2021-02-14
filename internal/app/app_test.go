package app

import "testing"

type HTMLDocuments struct {
	Data 	string
	Result 	string
}

func TestParseFunction(t *testing.T) {
	htmlDocuments := []HTMLDocuments{
		{
			Data: `<!DOCTYPE html>
					<html>
						<head>
							<title>First</title>
						</head>
						<body>
							<h1>My First Heading</h1>
							<p>My first paragraph.</p>
						</body>
					</html>`,
			Result: "First",
		},
		{
			Data: `<!DOCTYPE html>
					<html>
						<head>
							title>Second</title>
						</head>
						<body>
							<h1>My First Heading</h1>
							<p>My first paragraph.</p>
						</body>
					</html>`,
			Result: "",
		},
		{
			Data: `<!DOCTYPE html>
					<html>
						<body>
							<h1>My First Heading</h1>
							<p>My first paragraph.</p>
						</body>
					</html>`,
			Result: "",
		},
		{
			Data: `<!DOCTYPE html>
					<html>
						<body>
                            <title>bad_position</title>
							<h1>My First Heading</h1>
							<p>My first paragraph.</p>
						</body>
					</html>`,
			Result: "bad_position",
		},
	}

	for caseID, htmlDocument := range htmlDocuments {
		result, err := parseFunction([]byte(htmlDocument.Data))
		if err != nil {
			t.Error(err)
		}

		if result != htmlDocument.Result {
			t.Error("For case ", caseID, "expected result", htmlDocument.Result, "got", result)
		}
	}
}
