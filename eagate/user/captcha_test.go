package user

/*
func TestSolveCaptcha(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network-dependent tests in short mode")
	}
	// Setup test
	captchaDataMap := make(map[string]Captcha)
	const testDirectory = "./test_data"
	files, err := ioutil.ReadDir(testDirectory)
	if err != nil {
		t.Fatalf("could not read test_data dir")
	}
	if len(files) == 0 {
		t.Fatalf("no test files found")
	}
	for _, f := range files {
		fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", testDirectory, f.Name()))
		if err != nil {
			t.Fatalf("could not read test file %s: %s", f.Name(), err.Error())
		}
		captchaData := Captcha{}
		err = json.Unmarshal(fileData, &captchaData)
		if err != nil {
			t.Fatalf("could not unmarshal test file %s: %s", f.Name(), err.Error())
		}
		captchaDataMap[f.Name()] = captchaData
	}

	// Setup expected results
	type Result struct {
		Session string
		Correct string
		Err 	error
	}
	expectedResults := make(map[string]Result)

	expectedResults["captcha_success_1.json"] = Result{
		Session: "75124830025950619372710398870271",
		Correct: "__477afc966db0f8061070e0107d995fb5___a9e4ecff6430f71c1542ee6edf358813",
		Err:     nil,
	}
	expectedResults["captcha_success_2.json"] = Result{
		Session: "74394428884344336884070791384980",
		Correct: "_00ad02f3d0ef91e2bff2cf23e000fde5__e31fdc6cdb18287c9ed35b7dd90693de__",
		Err:     nil,
	}
	expectedResults["captcha_success_3.json"] = Result{
		Session: "65652918467415196901423548351436",
		Correct: "__8b759ebb7065b777a28784069415aba7__1f43fbf9aabcbf40f8a6e263f3911c42_",
		Err:     nil,
	}
	expectedResults["captcha_success_4.json"] = Result{
		Session: "48157496743337867251908826719843",
		Correct: "__27c85087a8494b2ce0a5a24dfd31e0db_7caffbd50e5aac2aa7fab9477d0c5ae9__",
		Err:     nil,
	}
	expectedResults["captcha_success_5.json"] = Result{
		Session: "39394473268542550479484767900017",
		Correct: "__7bcbddeda724399d9f4bd2ad9a015fd1___036088d83dba324ffc4f692ac4b33242",
		Err:     nil,
	}

	// Run test
	for k, v := range captchaDataMap {
		session, correct, err := SolveCaptcha(v)
		actualResult := Result{
			session,
			correct,
			err,
		}
		if actualResult != expectedResults[k] {
			t.Errorf("captcha result failed for test %s: expected %+#v got %+#v", k, expectedResults[k], actualResult)
		}
	}
}*/
