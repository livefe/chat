package com.nearby.chat.feature.map

import com.nearby.chat.BuildConfig
import org.osmdroid.tileprovider.tilesource.OnlineTileSourceBase
import org.osmdroid.tileprovider.tilesource.XYTileSource
import org.osmdroid.util.MapTileIndex

object TiandituTileSources {
    private fun hosts(pathPrefix: String): Array<String> {
        val tk = BuildConfig.TIANDITU_TK
        return Array(8) { i ->
            "https://t$i.tianditu.gov.cn/DataServer?T=$pathPrefix&tk=$tk"
        }
    }

    fun vectorBase(): OnlineTileSourceBase = object : XYTileSource(
        "TiandituVec",
        1,
        18,
        256,
        "",
        hosts("vec_w"),
    ) {
        override fun getTileURLString(pMapTileIndex: Long): String = baseUrl +
            "&x=${MapTileIndex.getX(pMapTileIndex)}" +
            "&y=${MapTileIndex.getY(pMapTileIndex)}" +
            "&l=${MapTileIndex.getZoom(pMapTileIndex)}"
    }

    fun vectorAnno(): OnlineTileSourceBase = object : XYTileSource(
        "TiandituCva",
        1,
        18,
        256,
        "",
        hosts("cva_w"),
    ) {
        override fun getTileURLString(pMapTileIndex: Long): String = baseUrl +
            "&x=${MapTileIndex.getX(pMapTileIndex)}" +
            "&y=${MapTileIndex.getY(pMapTileIndex)}" +
            "&l=${MapTileIndex.getZoom(pMapTileIndex)}"
    }
}
