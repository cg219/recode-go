import { createApp } from "https://unpkg.com/vue@3.4.21/dist/vue.esm-browser.js";
import NewEpisode from "./new-episode.js"
import RecodeQueue from "./recode-queue.js"
import NextEpisode from "./next-episode.js"
import NewShow from "./new-show.js"

createApp({
    data()  {
        return {
            title: "Recode GUI",
            list: [],
            rootdir: ""
        }
    },
    components: {
        NewEpisode,
        RecodeQueue,
        NextEpisode,
        NewShow,
    },
    async mounted() {
        try {
            const [list, rootdir] = await Promise.all([fetch("/anime").then((d) => d.json()), fetch("/rootdirectory", { method: "POST" }).then((d) => d.text())])

            this.list = list ?? []
            this.rootdir = rootdir ?? ""

        }  catch (err) {
            console.log(err)
        }
    }
}).mount("#app")
