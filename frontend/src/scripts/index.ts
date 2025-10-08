import {render} from "lit";

import "@altshiftab/styles/common_web.css";
import "@altshiftab/styles/common_header_footer.css";
import {addErrorEventListeners} from "@altshiftab/http_service_utils_js";
import "@altshiftab/web_components/button";
import "@altshiftab/web_components/footer"
import "@altshiftab/web_components/header"

import config from "../../../config.json";
type Config = typeof config

addErrorEventListeners();

const routeNames = Object.keys(config.routes);
const routePaths = ["/", ...Object.values(config.routes)];

async function importPage(name: string) {
    // TODO: Inspect the Webpack code responsible for this?
    // IMPORTANT: keep a static prefix ("./pages/") so webpack builds a context
    return import(
        /* webpackMode: "lazy", webpackChunkName: "page-[request]" */
        `./pages/${name}.ts`
    );
}

async function renderSpa(path = location.pathname) {
    const main = document.querySelector("main");
    if (!(main instanceof HTMLElement))
        throw new Error("no main element found");

    let name = "";
    if (path === "/") {
        name = "root";
    } else {
        const segments = path.split("/").filter(Boolean);
        if (segments.length !== 1)
            throw new Error("The path is of an unexpected length.");

        const segment = segments[0];
        if (!routeNames.includes(segment))
            throw new Error("The path is not a valid route.");

        name = segment;
    }

    const mod = await importPage(name);
    render(new mod.default(), main);

    document.body.classList.remove("loading");
}

addEventListener("click", (event: MouseEvent) => {
    if (event.defaultPrevented || event.button !== 0)
        return;

    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
        return;

    // Walk the composed path to support anchors in shadow DOM
    let anchor: HTMLAnchorElement | null = null;
    for (const node of event.composedPath()) {
        if (!(node instanceof Node))
            continue;

        if (node instanceof HTMLAnchorElement && node.hasAttribute("href")) {
            anchor = node;
            break;
        }

        const root = node.getRootNode() as Document | ShadowRoot;
        const scopedElement = node instanceof Element
            ? node
            : (root as ShadowRoot).host ?? null
        ;

        if (scopedElement && scopedElement instanceof Element) {
            const found = scopedElement.closest
                ? scopedElement.closest("a[href]")
                : null
            ;
            if (found) {
                anchor = found as HTMLAnchorElement;
                break;
            }
        }
    }

    if (!anchor)
        return;

    if (anchor.target !== "" && anchor.target.toLowerCase() !== "_self")
        return;

    if (anchor.hasAttribute("download"))
        return;

    const url = new URL(anchor.href, location.href);

    // External links
    if (url.origin !== location.origin)
        return;

    // Non-SPA routes
    if (!routePaths.includes(url.pathname))
        return;

    event.preventDefault();

    const hyperlinkDestination = url.pathname + url.search + url.hash
    const currentRelativeReference = location.pathname + location.search + location.hash;

    if (hyperlinkDestination === currentRelativeReference)
        return;

    const [pathAndSearch, hash = ""] = hyperlinkDestination.split("#", 2);
    if ((pathAndSearch || "") === (location.pathname + location.search)) {
        // Let the browser handle scrolling to anchors without re-render
        return void history.pushState(null, "", hyperlinkDestination);
    }

    history.pushState(null, "", hyperlinkDestination);

    // Ensure we start at the top for real navigations (no fragment)
    if (!hash)
        window.scrollTo({top: 0, left: 0, behavior: "auto"});

    renderSpa();
});

addEventListener("popstate", () => renderSpa())
addEventListener("DOMContentLoaded", () => {
    renderSpa();
});
