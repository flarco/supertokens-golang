/* Copyright (c) 2021, VRAI Labs and/or its affiliates. All rights reserved.
 *
 * This software is licensed under the Apache License, Version 2.0 (the
 * "License") as published by the Apache Software Foundation.
 *
 * You may not use this file except in compliance with the License. You may
 * obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package api

import (
	"encoding/json"
	"io/ioutil"

	"github.com/supertokens/supertokens-golang/recipe/emailpassword/epmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
)

func SignInAPI(apiImplementation epmodels.APIInterface, options epmodels.APIOptions) error {
	if apiImplementation.SignInPOST == nil || (*apiImplementation.SignInPOST) == nil {
		options.OtherHandler(options.Res, options.Req)
		return nil
	}

	body, err := ioutil.ReadAll(options.Req.Body)
	if err != nil {
		return err
	}
	var formFieldsRaw map[string]interface{}
	err = json.Unmarshal(body, &formFieldsRaw)
	if err != nil {
		return err
	}

	formFields, err := validateFormFieldsOrThrowError(options.Config.SignInFeature.FormFields, formFieldsRaw["formFields"].([]interface{}))
	if err != nil {
		return err
	}

	result, err := (*apiImplementation.SignInPOST)(formFields, options, &map[string]interface{}{})
	if err != nil {
		return err
	}
	if result.WrongCredentialsError != nil {
		return supertokens.Send200Response(options.Res, map[string]interface{}{
			"status": "WRONG_CREDENTIALS_ERROR",
		})
	} else {
		return supertokens.Send200Response(options.Res, map[string]interface{}{
			"status": "OK",
			"user":   result.OK.User,
		})
	}
}
