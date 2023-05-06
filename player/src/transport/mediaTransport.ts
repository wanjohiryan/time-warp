import * as Message from "./message"
import * as Stream from "../stream"
import * as MP4 from "../mp4"

import Audio from "../audio"
import Video from "../video"
import getUrl from "./getUrl"

export interface TransportInit {
	path: string;
	getLatency: (t: number) => void;

	audio: Audio;
	video: Video;
	// setLatency: (latency: number) => void;
}

export default class MediaTransport {
	quic: Promise<WebTransport>;
	api: Promise<WritableStream>;
	tracks: Map<string, MP4.InitParser>;
	getLatency: (t: number) => void;

	audio: Audio;
	video: Video;

	constructor(props: TransportInit) {
		this.tracks = new Map();

		this.audio = props.audio;
		this.video = props.video;
		this.getLatency = props.getLatency;

		this.quic = this.connect(props)

		// Create a unidirectional stream for all of our messages
		this.api = this.quic.then((q) => {
			return q.createUnidirectionalStream()
		})

		// async functions
		this.receiveStreams()
	}

	async close() {
		(await this.quic).close()
	}

	// Helper function to make creating a promise easier
	private async connect(props: TransportInit): Promise<WebTransport> {
		const url = getUrl(props.path)

		const quic = new WebTransport(url)
		await quic.ready
		return quic
	}

	async sendMessage(msg: any) {
		const payload = JSON.stringify(msg)
		const size = payload.length + 8

		const stream = await this.api

		const writer = new Stream.Writer(stream)
		await writer.uint32(size)
		await writer.string("warp")
		await writer.string(payload)
		writer.release()
	}

	async receiveStreams() {
		const q = await this.quic
		const streams = q.incomingUnidirectionalStreams.getReader()

		while (true) {
			const result = await streams.read()
			if (result.done) break

			const stream = result.value
			this.handleStream(stream) // don't await
		}
	}

	async handleStream(stream: ReadableStream) {
		let r = new Stream.Reader(stream)

		while (!await r.done()) {
			const size = await r.uint32();
			const typ = new TextDecoder('utf-8').decode(await r.bytes(4));

			if (typ != "warp") throw "expected warp atom"
			if (size < 8) throw "atom too small"

			const payload = new TextDecoder('utf-8').decode(await r.bytes(size - 8));
			const msg = JSON.parse(payload)

			if (msg.init) {
				return this.handleInit(r, msg.init as Message.Init)
			} else if (msg.segment) {
				return this.handleSegment(r, msg.segment as Message.Segment)
			} else if (msg.beat) {
				return this.handleHeartBeat(r, msg.beat as Message.Beat)
			}

		}
	}

	async handleInit(stream: Stream.Reader, msg: Message.Init) {
		let track = this.tracks.get(msg.id);
		if (!track) {
			track = new MP4.InitParser()
			this.tracks.set(msg.id, track)
		}

		while (1) {
			const data = await stream.read()
			if (!data) break

			track.push(data)
		}

		const info = await track.info

		if (info.audioTracks.length + info.videoTracks.length != 1) {
			throw new Error("expected a single track")
		}

		if (info.audioTracks.length) {
			this.audio.init({
				track: msg.id,
				info: info,
				raw: track.raw,
			})
		} else if (info.videoTracks.length) {
			this.video.init({
				track: msg.id,
				info: info,
				raw: track.raw,
			})
		} else {
			throw new Error("init is neither audio nor video")
		}
	}

	async handleSegment(stream: Stream.Reader, msg: Message.Segment) {
		let track = this.tracks.get(msg.init);
		if (!track) {
			track = new MP4.InitParser()
			this.tracks.set(msg.init, track)
		}

		// Wait until we learn if this is an audio or video track
		const info = await track.info

		if (info.audioTracks.length) {
			this.audio.segment({
				track: msg.init,
				buffer: stream.buffer,
				reader: stream.reader,
			})
		} else if (info.videoTracks.length) {
			this.video.segment({
				track: msg.init,
				buffer: stream.buffer,
				reader: stream.reader,
			})
		} else {
			throw new Error("segment is neither audio nor video")
		}
	}

	async handleHeartBeat(stream: Stream.Reader, msg: Message.Beat) {
		//TODO: use the initial latency to calculate the network quality over time
		// const now = Date.now()
		// console.log("initial latency:", now - msg.timestamp);

		while (1) {
			const data = await stream.read()
			if (!data) break
			const t = new TextDecoder('utf-8').decode(data);

			// console.log("latency", Date.now() - Number(t))
			this.getLatency(Date.now() - Number(t))
		}
	}

	async setMaxBitrate(value: number) {
		await this.sendMessage({
			"bandwidth": {
				"max_bitrate": value,
			}
		})
	}
}
