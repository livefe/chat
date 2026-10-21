package com.nearby.chat.data.push

import android.content.Context
import android.os.Handler
import android.os.Looper
import android.util.Log
import com.nearby.chat.BuildConfig
import com.xiaomi.mipush.sdk.MiPushClient
import com.xiaomi.mipush.sdk.Region

class XiaomiPushChannel(
    private val appContext: Context,
    private val tokenStore: PushTokenStore,
) : PushChannel {
    override val vendor = PushVendor.XIAOMI

    override fun getToken(onToken: TokenListener) {
        val appId = BuildConfig.XIAOMI_APP_ID
        val appKey = BuildConfig.XIAOMI_APP_KEY
        if (appId.isBlank() || appKey.isBlank()) {
            Log.w(TAG, "XIAOMI_APP_ID/KEY 未配置，跳过真通道注册")
            onToken.onToken("")
            return
        }

        tokenStore.setPendingListener(onToken)
        try {
            MiPushClient.setRegion(Region.China)
            MiPushClient.registerPush(appContext, appId, appKey)

            val existing = MiPushClient.getRegId(appContext)
            if (!existing.isNullOrBlank()) {
                Handler(Looper.getMainLooper()).post {
                    tokenStore.deliver(existing)
                }
            }
        } catch (t: Throwable) {
            Log.e(TAG, "MiPush register failed", t)
            onToken.onToken("")
        }
    }

    companion object {
        private const val TAG = "XiaomiPush"
    }
}
