package com.nearby.chat.data.push

import android.content.Context
import android.util.Log
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PushManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val tokenStore: PushTokenStore,
) {
    val vendor: PushVendor = VendorDetector.detect()

    private val channel: PushChannel = when (vendor) {
        PushVendor.XIAOMI -> XiaomiPushChannel(context.applicationContext, tokenStore)
        else -> StubPushChannel(vendor)
    }

    fun init(onToken: TokenListener) {
        Log.i(TAG, "init push vendor=$vendor")
        channel.getToken { token ->
            if (token.isBlank()) {
                Log.w(TAG, "empty push token vendor=$vendor")
                return@getToken
            }

            if (token.startsWith("stub-")) {
                Log.w(TAG, "skip registering stub push token vendor=$vendor")
                return@getToken
            }
            Log.i(TAG, "push token ready vendor=$vendor")
            onToken.onToken(token)
        }
    }

    companion object {
        private const val TAG = "PushManager"
    }
}
