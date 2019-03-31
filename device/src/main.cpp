#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <IRremoteESP8266.h>
#include <IRrecv.h>
#include <IRsend.h>
#include <IRutils.h>
#include <Device.pb.h>

int RECV_PIN = 14;
int SEND_PIN = 12;
IRrecv irrecv(RECV_PIN);
IRsend irsend(SEND_PIN);
decode_results results;

int status = WL_IDLE_STATUS;

void printCurrentNet() {
  Serial.print("SSID: ");
  Serial.println(WiFi.SSID());

  Serial.print("BSSID: ");
  Serial.println(WiFi.BSSIDstr());

  Serial.print("Signal Strength: ");
  long strength = WiFi.RSSI();
  Serial.println(strength);

}

void printMacAddress(byte mac[6]) {
  for (int i =5; i >= 0; i--) {
    Serial.print(mac[i], HEX);
    Serial.print(":");
  }
  Serial.println();
}

void printWifiData() {
  IPAddress ip = WiFi.localIP();
  Serial.print("IP: ");
  Serial.println(ip);

  byte mac[6];
  WiFi.macAddress(mac);
  Serial.print("MAC: ");
  printMacAddress(mac);
}

void setup()
{
  Serial.begin(9600);

  irsend.begin();

  //irrecv.enableIRIn();

  while(!Serial) {
  }

  Device_Command cmd = Device_Command_init_zero;


  status = WiFi.begin("fort-kickass", "william se wireless");
  Serial.println("Connecting");
  while(WiFi.status() != WL_CONNECTED) {
    Serial.print(".");
    delay(500);
  }

  Serial.println();
  Serial.println("Connected!");

  printCurrentNet();
  printWifiData();
}

void IRRecvProcessing() {
  if (irrecv.decode(&results)) {
    Serial.print("IR RECV Code = 0x ");
    serialPrintUint64(results.value, HEX);
    Serial.println();

    Serial.print("IR RAW Len: ");
    Serial.println(results.rawlen);
    for (int i = 0 ; i < results.rawlen; i++) {
      Serial.print(results.rawbuf[i]);
      Serial.print(" ");
    }
    Serial.println();

    irrecv.resume();
  }
}


void IRSendProcessing() {
  uint64_t JBL_PWR = 0x61FFD827UL;
  // send the samsung code twice quickly to ensure it is picked up
  uint64_t SAMSAUNG_PWR = 0xE0E040BFUL;
  Serial.println("TV on (SAMSUNG)");
  irsend.sendSAMSUNG(SAMSAUNG_PWR, 32, 2);
  delay(10000);
  Serial.println("Sound on (NEC)");
  irsend.sendNEC(JBL_PWR);
  delay(10000);
}

void loop()
{
  IRSendProcessing();
}