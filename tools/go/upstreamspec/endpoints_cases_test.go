package main

import "strings"

func positiveCases(specSha string) []endpointCase {
	return append(append(positiveBasicCases(specSha), positivePlantedCases(specSha)...), positiveWrapperCases(specSha)...)
}

func positiveWrapperCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:        "(6) wrapper method with Get prefix is not treated as request",
			code:        "func (c *Client) GetStrikes(ctx context.Context, page int) error { return nil }\nfunc (c *Client) W(ctx context.Context, page int) error { return c.GetStrikes(ctx, page) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 0,
			wantInSpec:  0,
		},
		{
			name:        "(7) accessor Get with single string argument is not treated as request",
			code:        "func (c *Client) H(ctx context.Context, h interface{ Get(string) string }) error { _ = h.Get(\"X-Pagination-Page\"); return nil }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 0,
			wantInSpec:  0,
		},
	}
}

func positiveBasicCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:        "(a) good path passes",
			code:        "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(d) Sprintf %d path matches {id}",
			code:        "func (c *Client) Fetch(ctx context.Context, id int) error { return c.Get(ctx, fmt.Sprintf(\"/items/%d\", id)) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(e) concat \"/x/\"+id matches",
			code:        "func (c *Client) Send(ctx context.Context, id string) error { return c.Post(ctx, \"/x/\" + id) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(f) query string stripped",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { return c.Get(ctx, \"/search?q=\" + q) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(j) func with Get(\"/a\") then Post(\"/b\")",
			code:        "func (c *Client) Fetch(ctx context.Context) error {\n\tif err := c.Get(ctx, \"/a\"); err != nil {\n\t\treturn err\n\t}\n\treturn c.Post(ctx, \"/b\")\n}\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 2,
			wantInSpec:  2,
		},
	}
}

func positivePlantedCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:        "(1) unused literal not reaching request call is ignored",
			code:        "func (c *Client) Fetch(ctx context.Context) error {\n\t_ = \"/ignored\"\n\treturn c.Get(ctx, \"/items\")\n}\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(4) embedded mid-segment param matches",
			code:        "func (c *Client) Fetch(ctx context.Context, id string) error { return c.Get(ctx, \"/items-\" + id) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(4b) embedded param in extension matches",
			code:        "func (c *Client) Send(ctx context.Context, id string) error { return c.Post(ctx, \"/x/\" + id + \".jpg\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(5) query concat cuts at question mark",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { return c.Get(ctx, \"/search?q=\" + q + \"&limit=10\") }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(5b) query suffix appended with += keeps the path resolved",
			code:        "func (c *Client) Search(ctx context.Context, q string) error { path := \"/search\"; if q != \"\" { path += \"?q=\" + q }; return c.Get(ctx, path) }\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  1,
		},
		{
			name:        "(8) allow entry matches module func method path",
			code:        "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n",
			allow:       "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2099-01-01\n",
			module:      "testmod",
			sha:         specSha,
			wantCodeOps: 1,
			wantInSpec:  0,
		},
	}
}

func negativeBasicCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(b) wrong path refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/wrong\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "/wrong not in sample v1.2.3 spec",
		},
		{
			name:     "(c) wrong verb refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Post(ctx, \"/items\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "POST /items not in sample v1.2.3 spec",
		},
		{
			name:     "(h) sha mismatch refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:   "testmod",
			sha:      strings.Repeat("0", 64),
			wantPart: "has sha256",
		},
		{
			name:     "(i) module with .. refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			module:   "../testmod",
			sha:      specSha,
			wantPart: "one clean path",
		},
		{
			name:     "(2) unresolvable ident path fails",
			code:     "func (c *Client) Fetch(ctx context.Context, p string) error { return c.Get(ctx, p) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
		{
			name:     "(2b) path segment appended with += fails",
			code:     "func (c *Client) Fetch(ctx context.Context) error { path := \"/search\"; path += \"/more\"; return c.Get(ctx, path) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
		{
			name:     "(3) dynamic prefix concat fails",
			code:     "func (c *Client) Fetch(ctx context.Context, p string) error { return c.Get(ctx, p + \"/items\") }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
	}
}

func negativeAllowCases(specSha string) []endpointCase {
	return append(negativeAllowValidationCases(specSha), negativeAllowMatchingCases(specSha)...)
}

func negativeAllowValidationCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(k) unknown key refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-01-01\n    unknown_key: bad\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "unknown_key",
		},
		{
			name:     "(l) duplicate entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test 1\n    expires: 2027-01-01\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: duplicate\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "duplicate",
		},
		{
			name:     "(m) missing reason refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "reason",
		},
		{
			name:     "(m2) missing method refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    path: /items\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "method is required",
		},
		{
			name:     "(m3) missing path refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path is required",
		},
	}
}

func negativeAllowMatchingCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(g) expired allow entry fails",
			code:     "func (c *Client) Old(ctx context.Context) error { return c.Delete(ctx, \"/not-in-spec\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Old\n    method: DELETE\n    path: /not-in-spec\n    reason: test\n    expires: 2020-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "DELETE /not-in-spec not in sample v1.2.3 spec",
		},
		{
			name:     "(n) malformed expiry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-13-45\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "expires",
		},
		{
			name:     "(o) stale entry refused",
			code:     "func (c *Client) Fetch(ctx context.Context) error { return c.Get(ctx, \"/items\") }\n",
			allow:    "allow:\n  - module: testmod\n    func: Client.Fetch\n    method: GET\n    path: /items\n    reason: test\n    expires: 2027-01-01\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "matches no off-spec call; remove it",
		},
	}
}

func negativeCases(specSha string) []endpointCase {
	return append(append(negativeBasicCases(specSha), negativeParamCases(specSha)...), negativeAllowCases(specSha)...)
}

func negativeParamCases(specSha string) []endpointCase {
	return []endpointCase{
		{
			name:     "(2c) param compound assign fails unresolved",
			code:     "func (c *Client) A(ctx context.Context, p string) error { p += \"/more\"; return c.Get(ctx, p) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
		{
			name:     "(2d) param conditional reassign fails unresolved",
			code:     "func (c *Client) B(ctx context.Context, p string) error { if p == \"\" { p = \"/default\" }; return c.Get(ctx, p) }\n",
			module:   "testmod",
			sha:      specSha,
			wantPart: "path unresolved",
		},
	}
}

func endpointTestCases(specSha string) []endpointCase {
	return append(positiveCases(specSha), negativeCases(specSha)...)
}
