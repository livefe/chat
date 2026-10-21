package com.nearby.chat.data.push

import android.content.Context
import android.util.Log
import com.nearby.chat.core.share.ShareLinks
import com.xiaomi.mipush.sdk.ErrorCode
import com.xiaomi.mipush.sdk.MiPushClient
import com.xiaomi.mipush.sdk.MiPushCommandMessage
import com.xiaomi.mipush.sdk.MiPushMessage
import com.xiaomi.mipush.sdk.PushMessageReceiver
import dagger.hilt.android.EntryPointAccessors
import org.json.JSONObject

class XiaomiPushReceiver : PushMessageReceiver() {
    override fun onReceiveRegisterResult(context: Context, message: MiPushCommandMessage) {
        val command = message.getCommand()
        if (MiPushClient.COMMAND_REGISTER == command) {
            if (message.getResultCode() == ErrorCode.SUCCESS) {
                val args = message.getCommandArguments()
                val regId = args?.firstOrNull().orEmpty()
                Log.i(TAG, "MiPush register ok regId len=${regId.length}")
                EntryPointAccessors.fromApplication(context.applicationContext, PushEntryPoint::class.java)
                    .pushTokenStore()
                    .deliver(regId)
            } else {
                Log.w(TAG, "MiPush register fail code=${message.getResultCode()} reason=${message.getReason()}")
            }
        }
    }

    override fun onCommandResult(context: Context, message: MiPushCommandMessage) {
        Log.d(TAG, "onCommandResult cmd=${message.getCommand()} code=${message.getResultCode()}")
    }

    override fun onNotificationMessageClicked(context: Context, message: MiPushMessage) {
        val peerId = message.getExtra()?.get(ShareLinks.EXTRA_PEER_ID)?.toLongOrNull()
            ?: peerIdFromContent(message.getContent())
        val intent = PushIntents.openPrivateChat(context, peerId) ?: return
        context.startActivity(intent)
    }

    private fun peerIdFromContent(content: String?): Long {
        if (content.isNullOrBlank()) return 0L
        return runCatching { JSONObject(content).optLong(ShareLinks.EXTRA_PEER_ID) }.getOrDefault(0L)
    }

    companion object {
        private const val TAG = "XiaomiPushReceiver"
    }
}
