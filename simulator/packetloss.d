import std.algorithm : map, sort, uniq;
import std.array : array, join;
import std.conv : to;
import std.format : format;
import std.getopt : config, getopt;
import std.process : executeShell;
import std.stdio : writefln, writeln;
import std.string : split, strip;

string helpstr = q"HELP
Oversimplified packet loss script.
This program intentionally has significant limitations. Use 'iptables' directly for advanced operations.

Remember to run this program with 'sudo'

Options:
Either long or short options are allowed
    --ports -p  <network ports> (comma-separated)
        The ports to apply packet loss to
    --name  -n  <executablename>
        Append ports used by any executables matching <executablename> to the ports list
    --rate  --probability  -r <rate> (floating-point value between 0 and 1, inclusive)
        The packet loss rate. Use 1 for "disconnect".
        Omitting this argument will set the rate to 0.0
    --flush -f
        Remove all packet loss rules

Examples:
    sudo packetloss -f
        Removes all packet loss rules, disabling packet loss

    sudo packetloss -p 12345,23456,34567 -r 0.25
        Applies 25% packet loss to ports 12345, 23456, and 34567

    sudo packetloss -n executablename -r 0.25
        Applies 25% packet loss to all ports used by all programs named "executablename"

    sudo packetloss -p 12345 -n executablename -r 0.25
        Also applies 25% packet loss to port 12345

    sudo packetloss -n executablename -f
        Lists ports used by "executablename", but does not apply packet loss
HELP";

ushort[] parsePorts(string csv) {
    ushort[] portsOut;
    if (csv.length == 0) {
        return portsOut;
    }
    foreach (part; csv.split(",")) {
        auto s = part.strip;
        if (s.length == 0) {
            continue;
        }
        portsOut ~= s.to!ushort;
    }
    return portsOut.sort.uniq.array;
}

void main(string[] args) {
    bool help = false;
    bool flush = false;
    string portsArg;
    string name;
    double prob = 0.0;

    string[] chains = ["INPUT", "OUTPUT"];
    string[] protocols = ["udp", "tcp"];
    ushort[] ports;

    auto argsCopy = args.dup;
    argsCopy.getopt(
        config.passThrough,
        "h|help", &help,
        "f|flush", &flush,
        "p|port|ports", &portsArg,
        "n|name", &name,
        "r|rate|probability", &prob
    );

    if (help || args.length == 1) {
        writeln(helpstr);
        return;
    }

    if (prob < 0.0 || prob > 1.0) {
        writeln("Error: rate/probability must be within [0.0, 1.0]");
        return;
    }

    ports ~= parsePorts(portsArg);

    if (name.length > 0) {
        auto pidCmd = format("pidof %s", name);
        auto pids = executeShell(pidCmd).output.split;
        writefln("%d pids matching program name '%s': %-(%s, %)", pids.length, name, pids);

        foreach (pid; pids) {
            auto cmd = format(
                `netstat -aputn | grep "0.0.0.0" | grep "%s" | awk '{print $4}' | awk -F ':' '{print $2}'`,
                pid
            );
            auto res = executeShell(cmd);
            if (res.output.length == 0) {
                continue;
            }
            ports ~= res.output
                .split
                .map!(a => a.strip.to!ushort)
                .array;
        }

        ports = ports.sort.uniq.array;
        writefln("Found ports: %(%d, %)", ports);
    }

    writeln("Flushing iptables chains...");
    foreach (chain; chains) {
        executeShell(format("iptables -F %s", chain));
    }

    if (!flush) {
        if (ports.length == 0) {
            writeln("No ports resolved. Nothing to apply.");
        } else {
            auto portCsv = ports.map!(p => to!string(p)).array.join(",");
            bool failed = false;

            outer:
            foreach (chain; chains) {
                foreach (proto; protocols) {
                    auto command = format(
                        "iptables -A %s -p %s -m multiport --destination-ports %s -m statistic --mode random --probability %.6f -j DROP",
                        chain, proto, portCsv, prob
                    );

                    writefln("Performing command:\n  %s", command);
                    auto r = executeShell(command);
                    if (r.status != 0) {
                        failed = true;
                        writefln("Error:\n%-(  %s\n%)", r.output.split("\n"));
                        break outer;
                    }
                }
            }

            if (failed) {
                writeln("Rule setup aborted due to command failure.");
            }
        }
    }

    writefln("\n\nResult of 'iptables -L':\n\n%-(  %s\n%)", executeShell("iptables -L").output.split("\n"));
}
