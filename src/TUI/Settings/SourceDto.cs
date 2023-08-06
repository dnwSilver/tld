using YamlDotNet.Serialization;


namespace TUI.Settings;

[YamlSerializable]
public class SourceDto
{
    [YamlMember]
    public string[] Tags { get; set; }

    [YamlMember]
    public string Name { get; set; }

    [YamlMember]
    public int ProjectId { get; set; } = 0;

    [YamlMember]
    public string Repo { get; set; }
}