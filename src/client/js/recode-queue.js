export default {
    data() {
        return {
            queue: [],
        }
    },
    props: {
        title: String
    },
    template: "#recode-queue",
    methods: {},
    async mounted() {
        try {
            const response = await fetch("/queue") 
            const queue = await response.json()

            this.queue = queue ?? []
        } catch (err) {
            console.log(err)
        }
    }
}
