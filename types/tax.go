/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

type (
	// TaxRate for receipt items
	TaxRate string
	// TaxSystem (sno) for receipt items
	TaxSystem string
)

const (
	TaxRateNone   TaxRate = "none"   // without VAT.
	TaxRateVat0   TaxRate = "vat0"   // VAT 0%.
	TaxRateVat10  TaxRate = "vat10"  // VAT 10%.
	TaxRateVat110 TaxRate = "vat110" // calculated rate 10/110.
	TaxRateVat20  TaxRate = "vat20"  // VAT 20%.
	TaxRateVat22  TaxRate = "vat22"  // VAT 22%.
	TaxRateVat120 TaxRate = "vat120" // calculated rate 20/120.
	TaxRateVat122 TaxRate = "vat122" // calculated rate 20/122.
	TaxRateVat5   TaxRate = "vat5"   // VAT 5%.
	TaxRateVat7   TaxRate = "vat7"   // calculated rate 22/122.
	TaxRateVat105 TaxRate = "vat105" // calculated rate 5/105.
	TaxRateVat107 TaxRate = "vat107" // calculated rate 7/107.

	TaxSystemOSN              TaxSystem = "osn"                // general taxation system.
	TaxSystemUSNIncome        TaxSystem = "usn_income"         // simplified taxation system (income).
	TaxSystemUSNIncomeOutcome TaxSystem = "usn_income_outcome" // simplified taxation system (income minus expenses).
	TaxSystemESN              TaxSystem = "esn"                // single agricultural tax.
	TaxSystemPatent           TaxSystem = "patent"             // patent taxation system.
)

func (r TaxRate) String() string {
	return string(r)
}

func (s TaxSystem) String() string {
	return string(s)
}
