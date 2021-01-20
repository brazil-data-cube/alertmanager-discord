local Message(text) = {
    name: name
};

local EmbeddedMessage(title, message, embeddedContent) = {
    title: title,
    message: message,
    embeddedContent: embeddedContent
};

{
    "title": {
        "firing": Message("%s There is a new prometheus alert that needs attention :fire:"),
        "resolved": Message("%s I bring you good news :partying_face: A problem alert generated on prometheus has been solved.")
    },
    "embedded": {
        "firing": EmbeddedMessage("New %s alert", "[Problem] %s on %s", "I am sending this message to inform you that Prometheus target %s is experiencing problems related to %s. The criticality of this event is classified as %s. Make sure everything is correct."),
        "resolved": EmbeddedMessage("%s resolved", "[Solved] %s on %s has been solved", "I am sending this message to inform you that the problem reported on Prometheus target %s has just been solved, and everything is now under control.")
    },
}
