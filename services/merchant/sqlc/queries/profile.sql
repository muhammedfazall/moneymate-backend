-- name: GetStoreProfileByStoreID :one
-- GetStoreProfileByStoreID retrieves the complete merchant profile including contact info, business details, and logo.
SELECT 
    id, owner_id, owner_name, contact_email, mobile_number,
    legal_name, COALESCE(dba_name, '') AS dba_name, business_type, COALESCE(tax_id, '') AS tax_id, registered_address,
    display_id, vpa, qr_code_base64, status::text, plan::text, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
FROM stores
WHERE id = sqlc.arg('store_id')
LIMIT 1;



-- name: UpdateStoreProfileByStoreID :one
-- UpdateStoreProfileByStoreID modifies merchant business info, contact details, and logo.
UPDATE stores
SET legal_name = COALESCE(NULLIF(sqlc.arg('legal_name')::text, ''), legal_name),
    dba_name = COALESCE(NULLIF(sqlc.arg('dba_name')::text, ''), dba_name),
    registered_address = COALESCE(NULLIF(sqlc.arg('registered_address')::text, ''), registered_address),
    business_type = COALESCE(NULLIF(sqlc.arg('business_type')::text, ''), business_type),
    tax_id = COALESCE(NULLIF(sqlc.arg('tax_id')::text, ''), tax_id),
    owner_name = COALESCE(NULLIF(sqlc.arg('owner_name')::text, ''), owner_name),
    contact_email = COALESCE(NULLIF(sqlc.arg('contact_email')::text, ''), contact_email),
    mobile_number = COALESCE(NULLIF(sqlc.arg('mobile_number')::text, ''), mobile_number),
    logo_url = COALESCE(NULLIF(sqlc.arg('logo_url')::text, ''), logo_url),
    updated_at = NOW()
WHERE id = sqlc.arg('store_id')
RETURNING id, owner_id, owner_name, contact_email, mobile_number,
    legal_name, COALESCE(dba_name, '') AS dba_name, business_type, COALESCE(tax_id, '') AS tax_id, registered_address,
    display_id, vpa, qr_code_base64, status::text, plan::text, COALESCE(logo_url, '') AS logo_url, created_at, updated_at;

