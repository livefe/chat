package com.nearby.chat.feature.map

import android.content.Context
import android.graphics.Color
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import com.nearby.chat.BuildConfig
import com.nearby.chat.core.theme.AppColors
import org.osmdroid.config.Configuration
import org.osmdroid.events.MapListener
import org.osmdroid.events.ScrollEvent
import org.osmdroid.events.ZoomEvent
import org.osmdroid.tileprovider.MapTileProviderBasic
import org.osmdroid.util.GeoPoint
import org.osmdroid.views.MapView
import org.osmdroid.views.overlay.TilesOverlay

@Composable
fun TiandituMapView(
    initial: MapCenter,
    modifier: Modifier = Modifier,
    moveTo: MapCenter? = null,
    onCenterChanged: (MapCenter) -> Unit = {},
) {
    val lifecycleOwner = LocalLifecycleOwner.current
    val centerCb = rememberUpdatedState(onCenterChanged)
    var mapViewRef by remember { mutableStateOf<MapView?>(null) }

    LaunchedEffect(moveTo, mapViewRef) {
        val target = moveTo ?: return@LaunchedEffect
        val map = mapViewRef ?: return@LaunchedEffect
        map.controller.setCenter(GeoPoint(target.lat, target.lng))
        target.zoom?.let { map.controller.setZoom(it.toDouble()) }
    }

    Box(modifier = modifier) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { ctx ->
                val prefs = ctx.getSharedPreferences("osmdroid", Context.MODE_PRIVATE)
                Configuration.getInstance().apply {
                    load(ctx, prefs)
                    userAgentValue = BuildConfig.APPLICATION_ID
                    osmdroidBasePath = ctx.cacheDir
                    osmdroidTileCache = ctx.cacheDir.resolve("osmdroid").also { it.mkdirs() }
                }

                MapView(ctx).apply {
                    setMultiTouchControls(true)
                    isTilesScaledToDpi = true

                    minZoomLevel = 1.0
                    maxZoomLevel = 18.0
                    setTileSource(TiandituTileSources.vectorBase())

                    val annoProvider = MapTileProviderBasic(ctx).apply {
                        tileSource = TiandituTileSources.vectorAnno()
                    }
                    overlays.add(
                        TilesOverlay(annoProvider, ctx).apply {
                            loadingBackgroundColor = Color.TRANSPARENT
                            loadingLineColor = Color.TRANSPARENT
                        },
                    )

                    controller.setZoom((initial.zoom ?: 16).toDouble())
                    controller.setCenter(GeoPoint(initial.lat, initial.lng))

                    addMapListener(object : MapListener {
                        override fun onScroll(event: ScrollEvent?): Boolean {
                            emitCenter(this@apply, centerCb.value)
                            return false
                        }

                        override fun onZoom(event: ZoomEvent?): Boolean {
                            emitCenter(this@apply, centerCb.value)
                            return false
                        }
                    })

                    post { emitCenter(this, centerCb.value) }
                    mapViewRef = this
                }
            },
            update = { },
        )

        Canvas(
            modifier = Modifier
                .fillMaxSize()
                .align(Alignment.Center),
        ) {
            val cx = size.width / 2f
            val cy = size.height / 2f
            val arm = 14.dp.toPx()
            val stroke = 2.dp.toPx()
            val color = AppColors.Brand
            drawLine(color, Offset(cx, cy - arm), Offset(cx, cy + arm), stroke, StrokeCap.Round)
            drawLine(color, Offset(cx - arm, cy), Offset(cx + arm, cy), stroke, StrokeCap.Round)
            drawCircle(color, radius = 2.5.dp.toPx(), center = Offset(cx, cy))
        }
    }

    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_RESUME -> mapViewRef?.onResume()
                Lifecycle.Event.ON_PAUSE -> mapViewRef?.onPause()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            mapViewRef?.onDetach()
            mapViewRef = null
        }
    }
}

private fun emitCenter(map: MapView, cb: (MapCenter) -> Unit) {
    val c = map.mapCenter ?: return
    cb(
        MapCenter(
            lat = c.latitude,
            lng = c.longitude,
            zoom = map.zoomLevelDouble.toInt(),
        ),
    )
}
