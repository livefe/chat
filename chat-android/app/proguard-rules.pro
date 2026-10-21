-keep class com.xiaomi.** { *; }
-keep class com.nearby.chat.data.push.XiaomiPushReceiver { *; }
-dontwarn com.xiaomi.**

-keepattributes *Annotation*, InnerClasses, EnclosingMethod, Signature, RuntimeVisibleAnnotations

-keep @kotlinx.serialization.Serializable class com.nearby.chat.data.remote.dto.** { *; }
-keepclassmembers class com.nearby.chat.data.remote.dto.** {
    <fields>;
    <init>(...);
}

-keep class com.nearby.chat.data.remote.dto.**$$serializer { *; }

-keep,allowobfuscation,allowshrinking interface retrofit2.Call
-keep,allowobfuscation,allowshrinking class retrofit2.Response
-keep,allowobfuscation,allowshrinking class kotlin.coroutines.Continuation

-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }
-keep class * extends dagger.hilt.android.internal.managers.ViewComponentManager$FragmentContextWrapper { *; }

-keepclasseswithmembers class * {
    @com.google.gson.annotations.SerializedName <fields>;
}

-keep class com.nearby.chat.BuildConfig { *; }

-keepclassmembers class androidx.compose.ui.platform.AndroidComposeView {
    androidx.compose.ui.semantics.SemanticsOwner getSemanticsOwner();
}
