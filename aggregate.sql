SELECT cc.id, GROUP_CONCAT(attachment_filename SEPARATOR ';') AS receipts, cc.running_number
FROM receipts r JOIN cost_claims cc ON cc.id = cost_claim_id
-- WHERE cost_claim_id IS NOT NULL AND MONTH(cc.created) = MONTH(CURDATE())
WHERE cost_claim_id IS NOT NULL AND MONTH(cc.created) IN (4,5,6)
GROUP BY cost_claim_id
ORDER BY cc.running_number;