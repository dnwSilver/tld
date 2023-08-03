using YamlDotNet.Serialization;
using YamlDotNet.Serialization.NamingConventions;


namespace TUI.Domain;

[YamlSerializable]
public class Settings
{
    public static Settings Init()
    {
        var deserializer = new DeserializerBuilder()
                          .WithNamingConvention(UnderscoredNamingConvention.Instance)
                          .Build();

        using var sr = new StreamReader("settings.yaml");
        return deserializer.Deserialize<Settings>(sr.ReadToEnd());
    }

    [YamlMember]
    public Project[] Projects { get; set; }
}