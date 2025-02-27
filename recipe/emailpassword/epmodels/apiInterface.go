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

package epmodels

import (
	"net/http"

	"github.com/supertokens/supertokens-golang/recipe/emailverification/evmodels"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
)

type APIOptions struct {
	RecipeImplementation                  RecipeInterface
	EmailVerificationRecipeImplementation evmodels.RecipeInterface
	Config                                TypeNormalisedInput
	RecipeID                              string
	Req                                   *http.Request
	Res                                   http.ResponseWriter
	OtherHandler                          http.HandlerFunc
}

type APIInterface struct {
	EmailExistsGET                 *func(email string, options APIOptions, userContext supertokens.UserContext) (EmailExistsGETResponse, error)
	GeneratePasswordResetTokenPOST *func(formFields []TypeFormField, options APIOptions, userContext supertokens.UserContext) (GeneratePasswordResetTokenPOSTResponse, error)
	PasswordResetPOST              *func(formFields []TypeFormField, token string, options APIOptions, userContext supertokens.UserContext) (ResetPasswordUsingTokenResponse, error)
	SignInPOST                     *func(formFields []TypeFormField, options APIOptions, userContext supertokens.UserContext) (SignInPOSTResponse, error)
	SignUpPOST                     *func(formFields []TypeFormField, options APIOptions, userContext supertokens.UserContext) (SignUpPOSTResponse, error)
}

type SignUpPOSTResponse struct {
	OK *struct {
		User    User
		Session sessmodels.SessionContainer
	}
	EmailAlreadyExistsError *struct{}
}

type SignInPOSTResponse struct {
	OK *struct {
		User    User
		Session sessmodels.SessionContainer
	}
	WrongCredentialsError *struct{}
}

type EmailExistsGETResponse struct {
	OK *struct{ Exists bool }
}

type GeneratePasswordResetTokenPOSTResponse struct {
	OK *struct{}
}
