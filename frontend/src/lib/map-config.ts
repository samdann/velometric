import type { LayerSpecification } from "maplibre-gl";

export const MAP_STYLE = "https://tiles.openfreemap.org/styles/liberty";

export const routeLayerSpec: LayerSpecification = {
  id: "route",
  type: "line",
  source: "route",
  layout: { "line-cap": "round", "line-join": "round" },
  paint: { "line-color": "#F97316", "line-width": 3, "line-opacity": 0.9 },
};
