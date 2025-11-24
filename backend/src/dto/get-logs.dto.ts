import { IsDateString } from 'class-validator';

export class GetLogsDto {
  @IsDateString()
  from: string;

  @IsDateString()
  to: string;
}
