# M-Pesa setup

The API now supports tenant-scoped Daraja STK Push and C2B configuration. Safaricom cannot call a laptop-only or localhost URL: deploy the API behind public HTTPS before testing callbacks.

## Deployment prerequisites

1. Set `POS_CREDENTIALS_KEY` to the output of `openssl rand -base64 32`. Keep the key stable and back it up securely; changing it makes stored provider credentials unreadable.
2. Set `POS_PUBLIC_URL` to the public HTTPS origin of the API, for example `https://pos.example.com` (no trailing slash).
3. Run the database migrations, including `004_mpesa_configuration.sql` and `005_payment_references_and_refunds.sql`.
4. Sign in as an owner/admin. Configure provider permissions as appropriate for any other operator who will manage payments.

## Configure a tenant

Call `PUT /api/v1/payments/providers/mpesa` with a bearer token and:

```json
{
  "environment": "sandbox",
  "consumer_key": "...",
  "consumer_secret": "...",
  "business_short_code": "...",
  "passkey": "..."
}
```

The API encrypts credentials in PostgreSQL and returns the tenant's tokenized STK callback, C2B validation, and C2B confirmation URLs. Treat the returned tokenized URLs as secrets. Configure C2B URLs with Safaricom using `POST /api/v1/payments/providers/mpesa/c2b/register`; Safaricom sandbox/production account setup and shortcode permissions still apply. STK callback URLs are sent with each STK request.

Checkout: create a sale with `payment_method: "mpesa_stk"` and `received_minor: 0`, then call `POST /api/v1/payments/intents` with the sale ID, amount in minor units, phone number, and an `Idempotency-Key`. Poll `GET /api/v1/payments/intents/{intentID}` until captured or failed. The sale becomes completed when its captured tenders cover the total.

For C2B, use the 12-character `payment_reference` returned when the sale is created as the bill reference for automatic matching. Unknown or mismatched receipts are retained in the unallocated queue and can be assigned through `POST /api/v1/payments/unallocated/{paymentID}/allocate`.

Do not use production credentials until the sandbox transaction, callback, duplicate-delivery, and reconciliation checks have passed. Reversal/B2C refunds require additional Daraja initiator/security-credential setup and are not enabled by this integration yet.
