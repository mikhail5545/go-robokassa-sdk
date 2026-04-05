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
	// PaymentObject for receipt items
	PaymentObject string
	// PaymentMethod for receipt items
	PaymentMethod string
)

const (
	PaymentObjectCommodity            PaymentObject = "commodity"             // Goods (except excisable goods)
	PaymentObjectExcise               PaymentObject = "excise"                // Excisable goods
	PaymentObjectJob                  PaymentObject = "job"                   // Job
	PaymentObjectService              PaymentObject = "service"               // Service
	PaymentObjectGamblingBet          PaymentObject = "gambling_bet"          // Gambling bet
	PaymentObjectGamblingPrize        PaymentObject = "gambling_prize"        // Gambling prize (winning a game of chance)
	PaymentObjectLottery              PaymentObject = "lottery"               // A lottery ticket or bet
	PaymentObjectLotteryWin           PaymentObject = "lottery_win"           // Lottery winnings
	PaymentObjectLotteryPrize         PaymentObject = "lottery_prize"         // Lottery winnings (docs alias)
	PaymentObjectIntellectualActivity PaymentObject = "intellectual_activity" // Provision of results of intellectual activity
	PaymentObjectPayment              PaymentObject = "payment"               // Payment (advance payment, deposit, loan, etc.)
	PaymentObjectAgentCommission      PaymentObject = "agent_commission"      // Agency fee
	PaymentObjectComposite            PaymentObject = "composite"             // A component of the calculation
	PaymentObjectResortFee            PaymentObject = "resort_fee"            // Resort fee
	PaymentObjectAnother              PaymentObject = "another"               // Another subject of calculation
	PaymentObjectPropertyRight        PaymentObject = "property_right"        // Property law
	PaymentObjectNonOperatingGain     PaymentObject = "non-operating_gain"    // Non-operating income
	PaymentObjectInsurancePremium     PaymentObject = "insurance_premium"     // Insurance premiums
	PaymentObjectSalesTax             PaymentObject = "sales_tax"             // Trade free
	PaymentObjectProductMark          PaymentObject = "tovar_mark"            // Marked goods with a marking code (except excisable goods)

	PaymentMethodFullPrepayment PaymentMethod = "full_prepayment" // Full prepayment before transfer of the item of payment.
	PaymentMethodFullPayment    PaymentMethod = "full_payment"    // Full payment at the time of transfer.
	PaymentMethodAdvance        PaymentMethod = "advance"         // Advance payment
	PaymentMethodPrepayment     PaymentMethod = "prepayment"      // Partial prepayment before the transfer of the item of payment.
	PaymentMethodPartialPayment PaymentMethod = "partial_payment" // Partial payment and credit
	PaymentMethodCredit         PaymentMethod = "credit"          // Transfer on credit with subsequent payment
	PaymentMethodCreditPayment  PaymentMethod = "credit_payment"  // Loan payment
)

func (m PaymentMethod) String() string {
	return string(m)
}

func (o PaymentObject) String() string {
	return string(o)
}
