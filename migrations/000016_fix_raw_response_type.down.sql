ALTER TABLE ai_results ALTER COLUMN raw_response TYPE JSONB USING 
    CASE 
        WHEN raw_response IS NULL THEN NULL 
        ELSE to_jsonb(raw_response::TEXT) 
    END;
