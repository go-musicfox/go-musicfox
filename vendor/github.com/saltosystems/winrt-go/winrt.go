package winrt

// common
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IClosable
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncOperation`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncOperationCompletedHandler`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.DateTime
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Size
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.TimeSpan
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Uri
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Rect
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.WwwFormUrlDecoder
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IStringable
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Numerics.Quaternion
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Numerics.Vector2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Numerics.Vector3
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Numerics.Matrix4x4
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IReference`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncAction
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncActionCompletedHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncOperationProgressHandler`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncOperationWithProgressCompletedHandler`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Point
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncOperationWithProgress`2

// advertisement
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcher -method-filter add_Received -method-filter remove_Received -method-filter add_Stopped -method-filter remove_Stopped -method-filter Start -method-filter Stop -method-filter get_Status -method-filter get_AllowExtendedAdvertisements -method-filter put_AllowExtendedAdvertisements -method-filter get_ScanningMode -method-filter put_ScanningMode -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementReceivedEventArgs -method-filter get_RawSignalStrengthInDBm -method-filter get_BluetoothAddress -method-filter get_Advertisement -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementWatcherStoppedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEManufacturerData
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisement -method-filter get_LocalName -method-filter put_LocalName -method-filter get_ServiceUuids -method-filter get_ManufacturerData -method-filter get_DataSections -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementDataSection -method-filter get_DataType -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementPublisher -method-filter get_Advertisement -method-filter Start -method-filter Stop -method-filter get_Status -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEAdvertisementPublisherStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.Advertisement.BluetoothLEScanningMode

// bluetooth
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothLEDevice -method-filter FromBluetoothAddressAsync -method-filter FromBluetoothAddressWithBluetoothAddressTypeAsync -method-filter Close -method-filter get_ConnectionStatus -method-filter add_ConnectionStatusChanged -method-filter remove_ConnectionStatusChanged -method-filter get_BluetoothDeviceId -method-filter GetGattServicesWithCacheModeAsync -method-filter GetGattServicesAsync -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothConnectionStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothAddressType
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothDeviceId -method-filter !FromId
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothCacheMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.BluetoothError

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattSession -method-filter FromDeviceIdAsync -method-filter get_MaintainConnection -method-filter put_MaintainConnection -method-filter get_CanMaintainConnection -method-filter Close -method-filter get_MaxPduSize -method-filter add_MaxPduSizeChanged -method-filter remove_MaxPduSizeChanged -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattDeviceServicesResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCommunicationStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattDeviceService -method-filter get_Uuid -method-filter Close -method-filter GetCharacteristicsAsync -method-filter GetCharacteristicsWithCacheModeAsync -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristicsResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristic -method-filter get_Uuid -method-filter get_CharacteristicProperties -method-filter WriteValueWithOptionAsync -method-filter WriteValueAsync -method-filter ReadValueWithCacheModeAsync -method-filter ReadValueAsync -method-filter WriteClientCharacteristicConfigurationDescriptorAsync -method-filter add_ValueChanged -method-filter remove_ValueChanged -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattCharacteristicProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattWriteOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattReadResult -method-filter !get_ProtocolError
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattClientCharacteristicConfigurationDescriptorValue
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Bluetooth.GenericAttributeProfile.GattValueChangedEventArgs

// event
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.TypedEventHandler`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.EventRegistrationToken

// buffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.Buffer -method-filter !CreateCopyFromMemoryBuffer -method-filter !CreateMemoryBufferOverIBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IDataReader -method-filter ReadBytes -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.DataReader -method-filter FromBuffer -method-filter ReadBytes -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IDataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.DataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter DataWriter -method-filter Close -method-filter !*

// vector
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IVector`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IVectorView`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IPropertySet
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.ValueSet
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IObservableVector`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IIterable`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IMapView`2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IIterator`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IVectorChangedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.CollectionChange
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.VectorChangedEventHandler`1

// media player
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlayer -method-filter !GetSurface
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.PlaybackMediaMarkerSequence
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlayerState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlayerSurface
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlayerAudioDeviceType
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.IMediaPlaybackSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlayerAudioCategory
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.StereoscopicVideoRenderMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaBreakManager
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackCommandManager
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackSession
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackSphericalVideoProjection
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackSessionOutputDegradationPolicyState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackSessionVideoConstrictionReason
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaBreak
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackCommandManagerCommandBehavior
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaCommandEnablingRule
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.SphericalVideoProjectionMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.PlaybackMediaMarker
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackItem
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackAudioTrackList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackVideoTrackList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaPlaybackTimedMetadataTrackList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaBreakSchedule
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.TimedMetadataTrackPresentationMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaItemDisplayProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.AutoLoadedDisplayPropertyKind
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Playback.MediaBreakInsertionMethod
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.StereoscopicVideoPackingMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.SphericalVideoFrameFormat
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.MediaRotation
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.MediaRatio
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.AudioEncodingProperties -method-filter !GetProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaProperties.IMediaEncodingProperties -method-filter !GetProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SystemMediaTransportControls
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaTimelineController
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaTimelineControllerState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SoundLevel
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaPlaybackStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaTimelineControllerState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SystemMediaTransportControlsDisplayUpdater
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaPlaybackAutoRepeatMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SystemMediaTransportControlsTimelineProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SystemMediaTransportControlsButtonPressedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.SystemMediaTransportControlsButton
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.ImageDisplayProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MediaPlaybackType
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.MusicDisplayProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.VideoDisplayProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.MediaCategory
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Render.AudioRenderCategory
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Devices.AudioDeviceRole
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Devices.VideoDeviceController -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Devices.AudioDeviceController
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Devices.IMediaDeviceController
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Devices.Core.CameraIntrinsics
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Audio.AudioStateMonitor
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Casting.CastingSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.IMediaSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MediaSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.ISingleSelectMediaTrackList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MediaSourceState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MediaStreamSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseStreamSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MediaBinder
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MediaStreamSourceErrorStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.IMediaStreamDescriptor
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseSourceBufferList
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseReadyState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseSourceBuffer
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseEndOfStreamStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Core.MseAppendMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.MediaProtectionManager
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.ServiceRequestedEventHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.RebootNeededEventHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.ComponentLoadFailedEventHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.ServiceRequestedEventArgs -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.ComponentLoadFailedEventArgs
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.RevocationAndRenewalInformation
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.MediaProtectionServiceCompletion
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Protection.IMediaProtectionServiceRequest
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Streaming.Adaptive.AdaptiveMediaSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Streaming.Adaptive.AdaptiveMediaSourceDiagnostics
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Streaming.Adaptive.AdaptiveMediaSourceCorrelatedTimes
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Streaming.Adaptive.AdaptiveMediaSourceAdvancedSettings
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameSource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameSourceInfo -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameSourceController
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameFormat
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.VideoMediaFrameFormat
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameSourceKind
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.MediaFrameSourceGroup
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.Frames.DepthMediaFrameFormat
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Media.Capture.MediaStreamType

// ui
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.UI.Color

// graphics
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Graphics.DirectX.Direct3D11.IDirect3DSurface
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Graphics.DirectX.Direct3D11.Direct3DSurfaceDescription
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Graphics.DirectX.Direct3D11.Direct3DMultisampleDescription
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Graphics.DirectX.DirectXPixelFormat
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Graphics.Effects.IGraphicsEffect

// storage
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFile
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFile2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageOpenOptions
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StreamedFileDataRequestedHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageFile -method-filter !GetThumbnailAsyncOverloadDefaultSizeDefaultOptions -method-filter !GetThumbnailAsyncOverloadDefaultOptions -method-filter !GetScaledImageAsThumbnailAsyncOverloadDefaultSizeDefaultOptions
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItem
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileAccessMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageDeleteOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileAttributes
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageItemTypes
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemProperties2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFolder
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItem2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageProvider
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StreamedFileDataRequest
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStreamedFileDataRequest
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StreamedFileFailureMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemPropertiesWithProvider
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFilePropertiesWithAvailability
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.StorageItemContentProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.ThumbnailMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.ThumbnailOptions
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.NameCollisionOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.CreationCollisionOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IRandomAccessStream
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.RandomAccessStreamReference -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IRandomAccessStreamReference
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IInputStream
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.InputStreamOptions
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IOutputStream
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IInputStreamReference
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.MusicProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.IStorageItemExtraProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.VideoProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.VideoOrientation

// devices
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Devices.Enumeration.DeviceInformation -method-filter !*

// networking
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.DownloadOperation
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundDownloadProgress
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundTransferCostPolicy
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.ResponseInformation
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundTransferPriority
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundTransferGroup
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundTransferStatus
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.BackgroundTransferBehavior
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.IBackgroundTransferOperation
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Networking.BackgroundTransfer.IBackgroundTransferOperationPriority

// web
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Web.Http.HttpClient -method-filter !*

// system
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.System.User -method-filter !*
