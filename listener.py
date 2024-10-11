import json
from pathlib import Path

from vosk import Model, KaldiRecognizer
import pyaudio

model = Model(str(Path("vosk-model-ja-0.22").resolve()))

# MIC initilize
recognizer = KaldiRecognizer(model, 16000)
mic = pyaudio.PyAudio()

# voskの設定
def engine():
    stream = mic.open(format=pyaudio.paInt16,
                       channels=1, 
                       rate=16000, 
                       input=True, 
                       frames_per_buffer=8192)
    # ストリーミングデータを読み取る
    while True:
        stream.start_stream()
        try:
            data = stream.read(4096)
            if recognizer.AcceptWaveform(data):
                result = recognizer.Result()
                # jsonに変換
                response_json = json.loads(result) 
                print("SYSTEM: ", response_json)
                response = response_json["text"].replace(" ","")
                return response
            else:
                pass
        except OSError:
            pass

listening = True

# listeningをループして音声認識
def bot_listen_hear():
    global listening
    
    # listeningループ
    while listening:
        response = engine()
        print("SYSTEM: ","-"*22, "なにか話しかけてください","-"*22)
        # 空白の場合はループを途中で抜ける
        if response.strip() == "":
            continue
        else:
            pass
        
        return response 

if __name__ == "__main__":
    try:
        while True:     
            # bot_listen_hear関数を実施してレスポンスを得る
            user_input = bot_listen_hear()
            print("USER: ",user_input)

    except KeyboardInterrupt:
        
        # ctrl+c でループ終了
        print("SYSTEM: Vosk終了")