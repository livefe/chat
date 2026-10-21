declare namespace AMap {
  type LngLatTuple = [number, number]

  interface LngLat {
    lng: number
    lat: number
    getLng(): number
    getLat(): number
  }

  class Pixel {
    constructor(x: number, y: number)
  }

  interface MapsEvent {
    lnglat: LngLat
  }

  abstract class Overlay {
    on(type: string, handler: (e: MapsEvent) => void): void
  }

  interface PathOverlayOptions {
    strokeColor?: string
    strokeWeight?: number
    strokeStyle?: 'solid' | 'dashed'
    fillColor?: string
    fillOpacity?: number
    zIndex?: number
    cursor?: string
  }

  interface CircleOptions extends PathOverlayOptions {
    center: LngLatTuple
    radius: number
  }

  class Circle extends Overlay {
    constructor(options: CircleOptions)
    setCenter(center: LngLatTuple): void
    setRadius(radius: number): void
  }

  interface PolygonOptions extends PathOverlayOptions {
    path: LngLatTuple[]
  }

  class Polygon extends Overlay {
    constructor(options: PolygonOptions)
    getPath(): LngLat[] | LngLat[][]
  }

  interface MarkerOptions {
    position: LngLatTuple
    content?: string
    offset?: Pixel
    zIndex?: number
  }

  class Marker extends Overlay {
    constructor(options: MarkerOptions)
    setPosition(position: LngLatTuple): void
  }

  class InfoWindow {
    constructor(options?: { offset?: Pixel })
    setContent(content: string): void
    open(map: Map, position: LngLatTuple): void
    close(): void
  }

  interface MapOptions {
    zoom?: number
    zooms?: [number, number]
    center?: LngLatTuple
    viewMode?: '2D' | '3D'
    dragEnable?: boolean
    zoomEnable?: boolean
  }

  class Map {
    constructor(container: HTMLElement, options?: MapOptions)
    on(type: 'click', handler: (e: MapsEvent) => void): void
    on(type: 'zoomend', handler: () => void): void
    add(overlays: Overlay | Overlay[]): void
    remove(overlays: Overlay | Overlay[]): void
    setCenter(center: LngLatTuple): void
    getZoom(): number
    setZoom(zoom: number): void
    setFitView(
      overlays: Overlay[],
      immediately?: boolean,
      avoid?: [number, number, number, number],
    ): void
    setStatus(status: { doubleClickZoom?: boolean }): void
    resize(): void
    destroy(): void
  }

  interface MouseToolDrawEvent {
    obj?: Polygon
  }

  class MouseTool {
    constructor(map: Map)
    polygon(options: PathOverlayOptions): void
    on(type: 'draw', handler: (e: MouseToolDrawEvent) => void): void
    off(type: 'draw', handler: (e: MouseToolDrawEvent) => void): void
    close(clearOverlays: boolean): void
  }

  interface PlaceSearchResult {
    poiList?: { pois?: { location: LngLat }[] }
  }

  class PlaceSearch {
    constructor(options?: { map?: Map })
    search(
      keyword: string,
      callback: (status: string, result: PlaceSearchResult | string) => void,
    ): void
  }
}
