import { Client } from "@heroiclabs/nakama-js";
import type { Session, Socket } from "@heroiclabs/nakama-js";

/**
 * Generates a random device ID in the format "XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX"
 * where X is a hexadecimal digit (0-9, A-F).
 * 
 * @returns A randomly generated device ID string
 */
function generateRandomDeviceId(): string {
    const generateHexSegment = (length: number): string => {
        let result = '';
        const hexChars = '0123456789ABCDEF';
        for (let i = 0; i < length; i++) {
            result += hexChars.charAt(Math.floor(Math.random() * hexChars.length));
        }
        return result;
    };

    return [
        generateHexSegment(8),
        generateHexSegment(8),
        generateHexSegment(8),
        generateHexSegment(8)
    ].join('-');
}


const client = new Client("defaultkey", "127.0.0.1", '7350');

const connections: Array<{session: Session, socket: Socket}> = [];
for (let i = 0; i < 2; i++) {
    const deviceId = generateRandomDeviceId();
    const create = true;
    const session = await client.authenticateDevice(deviceId, create, "jsclient "+deviceId);
    const socket = client.createSocket();

    const appearOnline = true;
    await socket.connect(session, appearOnline);
    
    socket.onmatchdata = async (matchData) => {
        const dataString = new TextDecoder().decode(matchData.data);
        const jsonData = JSON.parse(dataString);
        console.log(JSON.stringify(jsonData, null, 2));
    }

    socket.onmatchmakermatched = async (mmMatched) => {
        var match = await socket.joinMatch(mmMatched.match_id);
        console.log('matchmakermatched' + String(i), match)
        
        let y: string
        if (i == 0) {
            y = '-100'
        } else {
            y = '100'
        }
        socket.sendMatchState(match.match_id, 2, `{"FromNodeID":1,"Type":1,"Position":{"X":0,"Y":${y}}}`)
    };

    connections.push({
        session,
        socket,
    });
}

for (const conn of connections) {
    const mmTicket = await conn.socket.addMatchmaker("", 0, 0)

    console.log(conn.session.username, 'mmTicket', mmTicket);
}