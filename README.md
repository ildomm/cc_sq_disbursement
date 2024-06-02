[![Go Report Card](https://goreportcard.com/badge/github.com/ildomm/cc_sq_disbursement?cache=v1)](https://goreportcard.com/report/github.com/ildomm/cc_sq_disbursement)
![coverage](https://raw.githubusercontent.com/ildomm/cc_sq_disbursement/badges/.badges/main/coverage.svg)

# CC Disbursement
Automate the calculation of merchants’ disbursements payouts and 
commissions for existing, present in the CSV files, and new orders.

## Architecture

The application consists of two parts:
### 1. CSV orders files loader
- The loader is responsible for loading the CSV files into the database.
  - It is a background job that runs every 1 minute.

### 2. Order processor for disbursements
- The processor is responsible for processing the orders and calculating the disbursements.
- The processor has two execution modes:
    - Without parameters: process all the orders from the database when `created_at` = yesterday.
    - With date range parameters: process all the orders from the database when `created_at` is between the given dates.

### Database schema

```mermaid
---
title: Tables.
---
erDiagram
   orders ||--o{ merchants : "Belongs To"
   orders ||--o{ merchant_disbursements : "Contains"

   merchants }|..|{ orders : "One-to-Many"
   merchants }|..|{ merchant_disbursements : "One-to-Many"
           
```

## Build process
- Type `make build` to generate the binaries in the `bin` folder.

### Orders Loader
- The entry point is in `cmd/loader/main.go`

### Orders Processor
- The entry point is in `cmd/processor/main.go`

## Development

### Environment variables
Required environment variables for `loader` and `processor`: 
- `DATABASE_URL` - the Postgres database URL, eg: `postgres://user:password@host:port/database`

Optional environment variables only for `processor`:
- `ORDERS_FROM` - the start date to process the orders, eg: `2020-01-01`
- `ORDERS_TO` - the end date to process the orders, eg: `2020-01-01`

## Deployment
Steps to deploy the application:
1. Create a Postgres database
2. Set the `DATABASE_URL` environment variable. Use the Postgres database URL, eg: `postgres://user:password@host:port/database`
3. Build the binaries with `make build`
   - You will need to have Golang installed in your machine
   - The current version of Golang used is `1.21.4`
4. Upload the binaries to the server. Recommended to use /usr/local/bin
5. Create directories for the orders files. Recommended the following structure:
    - /usr/local/orders
    - /usr/local/orders/waiting
    - /usr/local/orders/imported
    - /usr/local/orders/failed
6. Configure the loader binary to run as background job. It does not require cron job configuration, because it relies on the internal scheduler. 
    - `nohup /usr/local/bin/loader > /dev/null 2>&1 &`  
7. Create a cron job to run the processor. It is recommended to run them every midnight.
    - `0 0 * * * /usr/local/bin/processor`

## Testing
### Local tests:
#### Unit tests:
Unit tests are all written with the standard go testing library.
- Type `make unit-test` to run them.

#### Tests coverage:
- Type `make coverage-report` to generate an HTML report with the test coverage.
- Type `make coverage-total` to check the total test coverage.

## Future improvements
- Structured logging
- Metrics/Monitoring
- Improve logs, the current ones are not very useful and just implemented as a PoC
- Import each file in a separate goroutine, and use a channel to communicate the results.
- Finish test coverage for the processor
- Tests coverage until reaches 90% of the code
- Process the daily disbursement step using db transactions
- Make the FeePercentage configurable and not hardcoded
- Make orders CSV file paths configurable and not hardcoded
- Improve database indexes
- The `orders` table could be partitioned by `created_at` to improve performance

## Missing features
- 90% test code coverage.
- Panic recovery handling

## Considerations
- The idea of splitting the processes in two binaries is to make them independent and scalable.
- If I could choose the deployment environment, I would use Kubernetes to deploy the application.
- I would use a CI/CD tool to automate the build and deployment process.
- I would use a tool to manage the environment variables, like Vault.
- On my understanding of the problem, the Orders attribute `created_at` should be datetime, not date.
- The disbursements are calculated always for every merchant, even if they have configured per-week or per-month disbursements. Doing so, it does standardize the process and makes it easier to maintain. 
In case any Merchant requires a different disbursement configuration, it would be easy to apply without the need for reprocessing the whole database.
- I have decided to keep the `fee correction` in a separate attribute in order to facilitate reporting.
