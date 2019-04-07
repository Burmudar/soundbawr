#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <IRremoteESP8266.h>
#include <IRrecv.h>
#include <IRsend.h>
#include <IRutils.h>
#include <Device.pb.h>
#include <pb_common.h>
#include <pb_decode.h>

int RECV_PIN = 14;
int SEND_PIN = 12;
IRrecv irrecv(RECV_PIN);
IRsend irsend(SEND_PIN);
decode_results results;

int status = WL_IDLE_STATUS;

WiFiServer server(30000);

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
  
  server.begin();
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

void processCommand(Device_Command* cmd) {
  if (cmd->device == Device_Command_DeviceType_SOUND_BAR && cmd->action == Device_Command_Action_TURN_ON) {
    uint64_t JBL_PWR = 0x61FFD827UL;
    for(int i = 0; i < 1; i++){
      Serial.printf("Sound on (NEC): %d \n", i);
      irsend.sendNEC(JBL_PWR);
      delay(500);
    }
  }

  if (cmd->device == Device_Command_DeviceType_TV && cmd->action == Device_Command_Action_TURN_ON) {
    uint64_t SAMSAUNG_PWR = 0xE0E040BFUL;
    Serial.println("TV on (SAMSUNG)");
    irsend.sendSAMSUNG(SAMSAUNG_PWR, 32, 2);
  }

}

bool callback(pb_istream_t *stream, uint8_t *buf, size_t count){
  WiFiClient* client = (WiFiClient*)stream->state;

  size_t bytesRead = client->readBytes(buf, count);
  Serial.printf("Read %d bytes", bytesRead);
  Serial.println();

  if (!client->connected() || bytesRead == 0 || bytesRead < count){
    stream->bytes_left = 0;
    return false;
  }

  return true;
}

Device_Command* handleClientRequest(WiFiClient* client) {
  Serial.println("Processing connected client!");
  Device_Command* cmd ;
  while(client->connected()) {
    uint8_t buff[Device_Command_size+1];
    
    uint read = client->readBytes(buff, Device_Command_size+1);

    Serial.printf("Read %d bytes\n", read);

    Device_Command msg = Device_Command_init_zero;

    pb_istream_t stream = /*{&callback, client, SIZE_MAX};*/ pb_istream_from_buffer(buff, Device_Command_size+1);
    pb_decode(&stream, Device_Command_fields, &msg);

    String device = "<NONE>";
    if (msg.device == Device_Command_DeviceType_TV) {
      device = "TV";
    } else if (msg.device == Device_Command_DeviceType_SOUND_BAR) {
      device = "SOUNDBAR";
    }

    String action = "<NONE>";
    if (msg.action == Device_Command_Action_TURN_ON) {
      action = "TURN_ON";
    } else if (msg.action == Device_Command_Action_TURN_OFF) {
      action = "TURN_OFF";
    }
    Serial.println("Device: " + device);
    Serial.println("Action: " + action);

    cmd = &msg;
  }
  Serial.println("Processing Done!");
  return cmd;
}

void loop()
{
  WiFiClient client = server.available();

  if (client) {
    Device_Command* cmd = handleClientRequest(&client);
    processCommand(cmd);
  }
}