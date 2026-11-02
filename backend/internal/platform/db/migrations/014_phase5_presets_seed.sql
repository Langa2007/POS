INSERT INTO capability_definitions (id, name, description, requires, conflicts, settings) VALUES
('batch_tracking', 'Batch Tracking', 'Track stock per batch/lot', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('expiry_tracking', 'Expiry Tracking', 'Track expiry dates and FEFO', '["batch_tracking"]'::jsonb, '[]'::jsonb, '[{"key":"inventory.pick_strategy","type":"select","default":"FEFO","options":["FEFO","FIFO"]}]'::jsonb),
('prescriptions', 'Prescriptions', 'Prescription capture and dispensing', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('controlled_substances', 'Controlled Substances', 'Register for controlled drugs', '["prescriptions"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('table_service', 'Table Service', 'Tables, tabs, and courses', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('kitchen_display', 'Kitchen Display', 'Order routing to kitchens', '["table_service"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('room_booking', 'Room Booking', 'Reservations and availability', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('guest_folio', 'Guest Folio', 'Charge to room accounts', '["room_booking"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('loyalty', 'Loyalty', 'Points and rewards', '[]'::jsonb, '[]'::jsonb, '[{"key":"loyalty.earn_rate","type":"number","default":1}]'::jsonb),
('credit_accounts', 'Credit Accounts', 'Sales on account', '[]'::jsonb, '[]'::jsonb, '[{"key":"credit.default_limit_minor","type":"number","default":0}]'::jsonb),
('multi_uom', 'Multiple Units of Measure', 'Units and conversions', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('weighed_items', 'Weighed Items', 'Scale integration and weight barcodes', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('quotes', 'Quotes', 'Quotes and proformas', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('delivery', 'Delivery', 'Delivery orders', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('serial_tracking', 'Serial Tracking', 'Track by serial/IMEI/VIN', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb)
ON CONFLICT (id) DO NOTHING;

INSERT INTO preset_definitions (id, name, description, data) VALUES
('retail', 'Retail', 'General retail/supermarket', '{"capabilities":["loyalty"],"settings":{"inventory.pick_strategy":"FIFO"},"roles":["owner","manager","cashier"]}'::jsonb),
('pharmacy', 'Pharmacy', 'Pharmacy with prescriptions and expiry', '{"capabilities":["batch_tracking","expiry_tracking","prescriptions","controlled_substances","multi_uom"],"settings":{"inventory.pick_strategy":"FEFO","sales.require_customer_for_categories":["controlled"]},"roles":["owner","pharmacist","dispenser","cashier"]}'::jsonb),
('restaurant', 'Restaurant', 'Restaurant with tables and kitchen', '{"capabilities":["table_service","kitchen_display"],"settings":{},"roles":["owner","manager","waiter","kitchen"]}'::jsonb),
('hospitality', 'Hospitality', 'Hotels and lodges', '{"capabilities":["room_booking","guest_folio"],"settings":{},"roles":["owner","reception","housekeeping"]}'::jsonb),
('hardware', 'Hardware', 'Hardware and building supplies', '{"capabilities":["multi_uom","quotes","credit_accounts"],"settings":{},"roles":["owner","manager","cashier"]}'::jsonb),
('salon', 'Salon', 'Salon and spa', '{"capabilities":["loyalty"],"settings":{},"roles":["owner","stylist","reception"]}'::jsonb)
ON CONFLICT (id) DO NOTHING;
