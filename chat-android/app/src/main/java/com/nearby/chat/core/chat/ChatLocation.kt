package com.nearby.chat.core.chat

import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.nearby.chat.R
import com.nearby.chat.core.format.PlaceNames
import com.nearby.chat.core.ui.UiText
import org.json.JSONObject

data class ChatLocationContent(
    val latitude: Double,
    val longitude: Double,
    val placeName: String,
)

fun ChatLocationContent.displayText(): UiText = if (placeName.isEmpty()) UiText.of(R.string.place_nearby) else UiText.of(placeName)

@Composable
fun ChatLocationContent.displayName(): String = placeName.ifEmpty { stringResource(R.string.place_nearby) }

fun encodeChatLocation(latitude: Double, longitude: Double, placeName: String): String = JSONObject()
    .put("latitude", latitude)
    .put("longitude", longitude)
    .put("placeName", PlaceNames.realOrEmpty(placeName).ifEmpty { PlaceNames.NEARBY_WIRE_VALUE })
    .toString()

fun parseChatLocation(content: String?): ChatLocationContent? {
    if (content.isNullOrBlank()) return null
    return runCatching {
        val o = JSONObject(content)
        val lat = (o.opt("latitude") as? Number)?.toDouble() ?: return null
        val lng = (o.opt("longitude") as? Number)?.toDouble() ?: return null
        if (lat !in -90.0..90.0 || lng !in -180.0..180.0) return null
        if (lat == 0.0 && lng == 0.0) return null
        ChatLocationContent(
            latitude = lat,
            longitude = lng,
            placeName = PlaceNames.realOrEmpty(o.optString("placeName")),
        )
    }.getOrNull()
}

fun openChatLocationMap(context: Context, loc: ChatLocationContent) {
    val lat = loc.latitude
    val lng = loc.longitude
    val name = loc.displayText().resolve(context)

    val amap = Intent(
        Intent.ACTION_VIEW,
        Uri.parse(
            "androidamap://viewMap?sourceApplication=${Uri.encode(context.getString(R.string.app_name))}" +
                "&poiname=${Uri.encode(name)}&lat=$lat&lon=$lng&dev=0",
        ),
    ).apply { setPackage("com.autonavi.minimap") }
    if (amap.resolveActivity(context.packageManager) != null) {
        context.startActivity(amap)
        return
    }

    val geo = Uri.parse("geo:$lat,$lng").buildUpon()
        .appendQueryParameter("q", "$name@$lat,$lng")
        .build()
    runCatching {
        context.startActivity(Intent(Intent.ACTION_VIEW, geo))
    }
}
