INSERT INTO assets(id,mall_id,code,version,payload)
VALUES ('asset-ext-001','mall-demo','F1-EXT-001',1,'{"ID":"asset-ext-001","MallID":"mall-demo","ZoneID":"zone-f1-a","Code":"F1-EXT-001","QRSecret":"demo-qr-001","Kind":"extinguisher","Status":"active","Version":1}')
ON CONFLICT (id) DO NOTHING;
INSERT INTO routes(id,mall_id,version,payload)
VALUES ('route-f1','mall-demo',1,'{"ID":"route-f1","MallID":"mall-demo","Name":"一层晨检路线","TeamID":"team-inspect","Stops":[{"AssetID":"asset-ext-001","Order":1,"Minutes":3}],"Version":1}')
ON CONFLICT (id) DO NOTHING;
