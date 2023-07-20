# flutter

A Flutter project for sfu-ws.

## Getting Started

- `cd sfu-ws/flutter`
- `flutter create --project-name flutter_sfu_wsx_example --org com.github.pion .`

## iOS

Add the following entry to your _Info.plist_ file, located in `./ios/Runner/Info.plist`:

```xml
<key>NSCameraUsageDescription</key>
<string>$(PRODUCT_NAME) Camera Usage!</string>
<key>NSMicrophoneUsageDescription</key>
<string>$(PRODUCT_NAME) Microphone Usage!</string>
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

Add the following to the bottom of the `./ios/Podfile`:

```ruby
# platform :ios, '13.0'
```

## Android

Add the following entry to your Android Manifest file, located in `./android/app/src/main/AndroidManifest.xml:

```xml
<uses-feature android:name="android.hardware.camera" />
<uses-feature android:name="android.hardware.camera.autofocus" />
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.CHANGE_NETWORK_STATE" />
<uses-permission android:name="android.permission.MODIFY_AUDIO_SETTINGS" />

<application android:usesCleartextTraffic="true" ...>
...
</application>
```

Edit`android/app/build.gradle`, modify minSdkVersion to 18

```gradle
    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId "com.github.pion.flutter_sfu_wsx_example"
        minSdkVersion 18 // <-- here
        targetSdkVersion 28
        versionCode flutterVersionCode.toInteger()
        versionName flutterVersionName
        testInstrumentationRunner "android.support.test.runner.AndroidJUnitRunner"
    }
```

## Run

- `flutter pub get`
- `flutter run`
