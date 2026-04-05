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

package items

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mikhail5545/go-robokassa-sdk/types"
)

func (r *ReceiptItem) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 128)),
		validation.Field(&r.Quantity, validation.Required, validation.Min(1)),
		validation.Field(&r.Sum, validation.NilOrNotEmpty, validation.Min(float64(1))),
		validation.Field(&r.Cost, validation.NilOrNotEmpty, validation.Min(float64(1))),
		validation.Field(&r.Tax, validation.Required, validation.In(
			types.TaxRateNone, types.TaxRateVat0, types.TaxRateVat5,
			types.TaxRateVat7, types.TaxRateVat10, types.TaxRateVat20,
			types.TaxRateVat22, types.TaxRateVat105, types.TaxRateVat107,
			types.TaxRateVat110, types.TaxRateVat120, types.TaxRateVat122,
		)),
		validation.Field(&r.PaymentMethod, validation.NilOrNotEmpty, validation.In(
			types.PaymentMethodCredit, types.PaymentMethodCreditPayment,
			types.PaymentMethodPartialPayment, types.PaymentMethodFullPrepayment,
			types.PaymentMethodFullPayment, types.PaymentMethodPrepayment,
			types.PaymentMethodAdvance,
		)),
		validation.Field(&r.PaymentObject, validation.NilOrNotEmpty, validation.In(
			types.PaymentObjectAnother, types.PaymentObjectCommodity, types.PaymentObjectComposite,
			types.PaymentObjectAgentCommission, types.PaymentObjectExcise, types.PaymentObjectJob,
			types.PaymentObjectGamblingPrize, types.PaymentObjectService, types.PaymentObjectGamblingBet,
			types.PaymentObjectLottery, types.PaymentObjectLotteryWin, types.PaymentObjectLotteryPrize,
			types.PaymentObjectIntellectualActivity, types.PaymentObjectPayment, types.PaymentObjectResortFee,
			types.PaymentObjectPropertyRight, types.PaymentObjectNonOperatingGain, types.PaymentObjectInsurancePremium,
			types.PaymentObjectSalesTax, types.PaymentObjectProductMark,
		)),
		validation.Field(&r.NomenclatureCode, validation.NilOrNotEmpty),
	)
}
