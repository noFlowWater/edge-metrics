import { type RouteConfig, index, route, layout } from "@react-router/dev/routes";

export default [
  layout("components/Layout.tsx", [
    index("routes/home.tsx"),
    route("devices", "routes/devices.tsx"),
    route("devices/new", "routes/devices.new.tsx"),
    route("devices/:id", "routes/devices.$id.tsx"),
    route("kubernetes", "routes/kubernetes.tsx"),
  ]),
] satisfies RouteConfig;
