export default {
    data() {
        return {
            selected: ""
        }
    },
    props: {
        list: Array,
    },
    template: "#anime-list",
    methods: {},
    watch: {
        selected(value) {
            this.$emit("value-changed", value)
        }
    }
}
