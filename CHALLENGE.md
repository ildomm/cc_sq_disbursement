Backend coding challenge
==============================================================================

This coding challenge is for people who applied to the Senior Backend Developer position at (censored). The problem to solve is a simplified version of our daily problems.

Context
-------

(censored) provides e-commerce shops with a flexible payment method that allows shoppers to split their purchases in three months without any cost. In exchange, (censored) earns a fee for each purchase.

When shoppers use this payment method, they pay directly to (censored). Then, (censored) disburses the orders to merchants with different frequencies and pricing.

This challenge is about implementing the process of paying merchants.

Problem statement
-----------------

We have to implement a system to automate the calculation of merchants’ disbursements payouts and (censored) commissions for existing, present in the CSV files, and new orders.

The system must comply with the following requirements:

*   All orders must be disbursed precisely once.
*   Each disbursement, the group of orders paid on the same date for a merchant, must have a unique alphanumerical `reference`.
*   Orders, amounts, and fees included in disbursements must be easily identifiable for reporting purposes.

The disbursements calculation process must be completed, for all merchants, by 8:00 UTC daily, only including those merchants that fulfill the requirements to be disbursed on that day. Merchants can be disbursed daily or weekly. We will make weekly disbursements on the same weekday as their `live_on` date (when the merchant started using (censored), present in the CSV files). Disbursements groups all the orders for a merchant in a given day or week.

For each order included in a disbursement, (censored) will take a commission, which will be subtracted from the merchant order value gross of the current disbursement, following this pricing:

*   `1.00 %` fee for orders with an amount strictly smaller than `50 €`.
*   `0.95 %` fee for orders with an amount between `50 €` and `300 €`.
*   `0.85 %` fee for orders with an amount of `300 €` or more.

_Remember that we are dealing with money, so we should be careful with related operations. In this case, we should round up to two decimals following._

Lastly, on the first disbursement of each month, we have to ensure the `minimum_monthly_fee` for the previous month was reached. The `minimum_monthly_fee` ensures that (censored) earns at least a given amount for each merchant.

When a merchant generates less than the `minimum_monthly_fee` of orders’ commissions in the previous month, we will charge the amount left, up to the `minimum_monthly_fee` configured, as “monthly fee”. Nothing will be charged if the merchant generated more fees than the `minimum_monthly_fee`.

Charging the `minimum_monthly_fee` is out of the scope of this challenge. It is not subtracted from the disbursement commissions. Just calculate and store it for later usage.

Data
----

### Merchants sample

    id                                   | REFERENCE                 | EMAIL                             | LIVE_ON    | DISBURSEMENT_FREQUENCY | MINIMUM_MONTHLY_FEE
    2ae89f6d-e210-4993-b4d1-0bd2d279da62 | treutel_schumm_fadel      | info@treutel-schumm-and-fadel.com | 2022-01-01 | WEEKLY                 | 29.0
    6596b87d-7f13-460f-ba1a-00872c770092 | windler_and_sons          | info@windler-and-sons.com         | 2021-05-25 | DAILY                  | 29.0
    70de4478-bfa8-4c4c-97f1-4a0a149f8264 | mraz_and_sons             | info@mraz-and-sons.com            | 2020-03-20 | WEEKLY                 |  0.0
    52f0e308-4a9d-4b32-ace4-c491f457d9a5 | cummerata_llc             | info@cummerata-llc.com            | 2019-02-04 | DAILY                  | 35.0


You can find [merchants CSV here](/samples/merchants.csv).

### Orders samples

    id           | MERCHANT REFERENCE      | AMOUNT | CREATED AT
    056d024481a9 | treutel_schumm_fadel    |  61.74 | 2023-01-01
    33c80364591c | cummerata_llc           | 293.08 | 2023-01-01
    5eaeabf54862 | mraz_and_sons           | 373.33 | 2023-01-01
    70530cdc7b59 | treutel_schumm_fadel    |  60.48 | 2023-01-01
    871e0d072782 | mraz_and_sons           | 213.97 | 2023-01-01


You can find [orders CSV here](/samples/orders.csv).

We expect you to:

*   Create the necessary data structures and a way to persist them for the provided data. You don’t have to follow CSV’s schema if you think another one suits you better.
*   Calculate and store the disbursements following described requirements for all the orders included in the CSV, and prepare the system to do the same for new orders.
