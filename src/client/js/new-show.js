import AnimeList from "./anime-list.js"

export default {
    data() {
        return {
            nextShow: ""
        }
    },
    components: {
        AnimeList
    },
    props: {
        title: String,
        list: Array,
        rootdir: String
    },
    template: "#new-show",
    methods: {
        async addNewEpisode() {
            const data = new FormData()

            data.append("video", this.newEpisode.video, this.newEpisode.video.name)
            data.append("season", this.newEpisode.season)
            data.append("episode", this.newEpisode.episode)
            data.append("newepisode", this.newEpisode.newepisode)

            await fetch("/newepisode",  {
                method: "POST",
                body: data
            })
        },
        handleSelectUpdate(v) {
            this.nextShow = v
        }
    },

}
