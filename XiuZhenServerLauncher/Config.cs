using System.Collections.Generic;

namespace XiuZhenServerLauncher
{
    class Config
    {
        public Ftp Ftp { get; set; }
        public List<string> Whitelist { get; set; }
        public List<Region> Regions { get; set; }
    }

    class Region
    {
        public int RegionId { get; set; }
        public string WorkDir { get; set; }
        public string Start { get; set; }
        public string Stop { get; set; }
    }

    class Ftp
    {
        public string Host { get; set; }
        public int Port { get; set; }
        public string User { get; set; }
        public string Password { get; set; }
        public string Path { get; set; }
    }
}
