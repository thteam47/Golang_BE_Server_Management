rm *.pem
#1. tao khoa rieng tu va chung chi so cho CA
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=VN/ST=Hai Duong/L=Hai Duong/O=THteaM/OU=THteaM/CN=*thteam.vn/emailAddress=thteam47@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

#2. tao khoa rieng tu cho may chu web muon su dung tls
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=VN/ST=Ha Noi/L=Thanh Tri/O=Golang/OU=Computer/CN=*thteam47.vn/emailAddress=kidothteam@gmail.com"

#3. dung khoa rieng cua CA de ky CSR cua may chu web va lay lai chung chi da ki
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf
# add days openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem
echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text
#-> cai openssl

#4. tao khoa rieng tu cho client muon su dung tls
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=VN/ST=Ha Noi/L=Ha Dong/O=GolangClient/OU=Computer/CN=*thteam47.com/emailAddress=kidoclient@gmail.com"

#5. dung khoa rieng cua CA de ky CSR cua client va lay lai chung chi da ki
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf
echo "CLient's signed certificate"
openssl x509 -in client-cert.pem -noout -text